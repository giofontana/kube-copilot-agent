#!/usr/bin/env python3
import asyncio
import json
import os
import re
import signal
import subprocess
import uuid
from pathlib import Path

import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="KubeCopilot Agent")

COPILOT_HOME = Path(os.environ.get("COPILOT_HOME", "/copilot"))
SESSIONS_DIR = COPILOT_HOME / "sessions"
SESSIONS_DIR.mkdir(parents=True, exist_ok=True)

WEBHOOK_URL = os.environ.get("WEBHOOK_URL", "")

# In-memory async queue for background processing
_queue: asyncio.Queue = asyncio.Queue()

# Track active copilot subprocesses: queue_id → subprocess.Popen
_active_procs: dict[str, subprocess.Popen] = {}

ANSI_ESCAPE = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')


class ChatRequest(BaseModel):
    message: str
    session_id: str | None = None


class AsyncChatRequest(BaseModel):
    message: str
    session_id: str | None = None
    # Optional metadata passed back in the webhook payload
    send_ref: str | None = None
    namespace: str | None = None
    agent_ref: str | None = None


class ChatResponse(BaseModel):
    response: str
    session_id: str


def load_session(session_id: str) -> list[dict]:
    path = SESSIONS_DIR / f"{session_id}.json"
    if path.exists():
        return json.loads(path.read_text())
    return []


def save_session(session_id: str, history: list[dict]) -> None:
    path = SESSIONS_DIR / f"{session_id}.json"
    path.write_text(json.dumps(history, indent=2))


def strip_ansi(text: str) -> str:
    return ANSI_ESCAPE.sub('', text).strip()


def parse_copilot_jsonl(raw: str) -> tuple[str, str | None]:
    """
    Parse JSONL output from copilot --output-format json.
    Returns (response_text, session_id).
    """
    response_content = ""
    session_id = None

    for line in raw.strip().split('\n'):
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
            event_type = obj.get('type', '')
            if event_type == 'assistant.message':
                response_content = obj['data']['content']
            elif event_type == 'result':
                session_id = obj.get('sessionId')
        except (json.JSONDecodeError, KeyError):
            continue

    return response_content or strip_ansi(raw), session_id


def run_copilot(message: str, session_id: str | None = None) -> tuple[str, str | None]:
    """
    Invoke the copilot binary and return (response_text, session_id).
    copilot-instructions.md and skills are loaded automatically from
    COPILOT_HOME via --config-dir. Skills must be in subdirectories
    under /copilot/skills/<skill-name>/SKILL.md (native agent skills format).
    """
    cmd = [
        "copilot",
        "--config-dir", str(COPILOT_HOME),
        "--output-format", "json",
        "--allow-all-tools",
        "-p", message,
    ]
    if session_id:
        cmd.append(f"--resume={session_id}")

    env = os.environ.copy()
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=120,
            env={**env, "GH_NO_UPDATE_NOTIFIER": "1", "NO_COLOR": "1"},
        )
        stdout = result.stdout.strip()
        stderr = strip_ansi(result.stderr).strip()

        if result.returncode != 0:
            detail = strip_ansi(stdout) or stderr or "copilot returned a non-zero exit code"
            raise HTTPException(
                status_code=502,
                detail=f"copilot error (exit {result.returncode}): {detail}",
            )

        return parse_copilot_jsonl(stdout or "No response from copilot")
    except HTTPException:
        raise
    except subprocess.TimeoutExpired:
        raise HTTPException(status_code=504, detail="Copilot CLI timed out")
    except FileNotFoundError:
        raise HTTPException(status_code=500, detail="copilot binary not found in PATH")
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


async def _process_queue():
    """Background worker: processes async chat requests one by one."""
    while True:
        item = await _queue.get()
        queue_id = item["queue_id"]
        message = item["message"]
        session_id = item.get("session_id")
        send_ref = item.get("send_ref")
        namespace = item.get("namespace")
        agent_ref = item.get("agent_ref")
        try:
            response_text, resolved_session_id = await run_copilot_streaming(
                message, session_id, send_ref, namespace, agent_ref, queue_id
            )

            history = load_session(resolved_session_id)
            history.append({"user": message, "assistant": response_text})
            save_session(resolved_session_id, history)

            if WEBHOOK_URL:
                payload = {
                    "queue_id": queue_id,
                    "session_id": resolved_session_id,
                    "prompt": message,
                    "response": response_text,
                    "send_ref": send_ref,
                    "namespace": namespace,
                    "agent_ref": agent_ref,
                }
                try:
                    async with httpx.AsyncClient(timeout=10.0) as client:
                        await client.post(WEBHOOK_URL, json=payload)
                except Exception as e:
                    print(f"[asyncchat] webhook POST failed for queue_id={queue_id}: {e}")
        except Exception as e:
            print(f"[asyncchat] processing failed for queue_id={queue_id}: {e}")
        finally:
            _active_procs.pop(queue_id, None)
            _queue.task_done()


async def _post_chunk(chunk_url: str, send_ref: str, session_id: str | None,
                      agent_ref: str | None, namespace: str | None,
                      sequence: int, chunk_type: str, content: str):
    """Fire-and-forget POST of a streaming chunk to the operator webhook."""
    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            await client.post(chunk_url, json={
                "send_ref": send_ref or "",
                "session_id": session_id or "",
                "agent_ref": agent_ref or "",
                "namespace": namespace or "",
                "sequence": sequence,
                "chunk_type": chunk_type,
                "content": content,
            })
    except Exception as e:
        print(f"[chunk] POST failed seq={sequence}: {e}")


SKIP_TOOLS = {"report_intent", "skill"}


def _event_to_chunk(event: dict, tool_names: dict[str, str]) -> tuple[str, str] | None:
    """
    Convert a copilot events.jsonl entry to a (chunk_type, content) pair.
    tool_names maps toolCallId → toolName (populated from execution_start events).
    Returns None if the event should be skipped.
    """
    t = event.get("type", "")
    d = event.get("data", {})

    if t == "assistant.message":
        reasoning = d.get("reasoningText", "").strip()
        content = d.get("content", "").strip()
        tool_requests = d.get("toolRequests", [])

        if reasoning:
            return ("thinking", f"🤔 {reasoning[:300]}")
        if tool_requests:
            names = ", ".join(tr.get("name", "?") for tr in tool_requests
                             if tr.get("name") not in SKIP_TOOLS)
            if not names:
                return None
            return ("tool_call", f"Invoking: **{names}**")
        if content:
            preview = content[:200] + ("…" if len(content) > 200 else "")
            return ("response", f"💬 {preview}")

    elif t == "tool.execution_start":
        tool_name = d.get("toolName", "?")
        # Record callId→name for use in completion event
        call_id = d.get("toolCallId", "")
        if call_id:
            tool_names[call_id] = tool_name
        if tool_name in SKIP_TOOLS:
            return None
        args = d.get("arguments", {})
        desc = args.get("description") or args.get("command") or str(args)[:120]
        return ("tool_call", f"🔧 **{tool_name}**: {desc[:200]}")

    elif t == "tool.execution_complete":
        call_id = d.get("toolCallId", "")
        tool_name = tool_names.get(call_id, "")
        if tool_name in SKIP_TOOLS:
            return None
        result = d.get("result", {})
        output = result.get("content", "") or result.get("detailedContent", "")
        if not output or not output.strip():
            return None
        success = "✅" if d.get("success", True) else "❌"
        label = f" **{tool_name}**" if tool_name else ""
        return ("tool_result", f"{success}{label} result:\n```\n{output[:400]}\n```")

    elif t == "skill.invoked":
        name = d.get("name", "?")
        return ("info", f"📚 Skill loaded: **{name}**")

    return None


async def run_copilot_streaming(
    message: str, session_id: str | None,
    send_ref: str | None, namespace: str | None, agent_ref: str | None,
    queue_id: str | None = None,
) -> tuple[str, str]:
    """
    Run copilot via Popen, tail the session events.jsonl in real-time,
    POST each meaningful event as a KubeCopilotChunk, return (response_text, session_id).
    """
    chunk_url = WEBHOOK_URL.replace("/response", "/chunk") if WEBHOOK_URL else ""
    session_state_dir = COPILOT_HOME / "session-state"
    session_state_dir.mkdir(parents=True, exist_ok=True)

    cmd = [
        "copilot",
        "--config-dir", str(COPILOT_HOME),
        "--output-format", "json",
        "--allow-all-tools",
        "-p", message,
    ]
    if session_id:
        cmd.append(f"--resume={session_id}")

    env = {**os.environ.copy(), "GH_NO_UPDATE_NOTIFIER": "1", "NO_COLOR": "1"}

    # Snapshot existing sessions before launch to detect the new one
    existing_sessions = {d.name for d in session_state_dir.iterdir() if d.is_dir()}

    loop = asyncio.get_event_loop()
    proc = await loop.run_in_executor(None, lambda: subprocess.Popen(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, env=env,
        start_new_session=True,  # new process group so we can kill all children
    ))

    # Register proc so /cancel can kill the entire process group
    if queue_id:
        _active_procs[queue_id] = proc

    sequence = 0
    if chunk_url and send_ref:
        await _post_chunk(chunk_url, send_ref, session_id, agent_ref, namespace,
                          sequence, "info", f"Processing: {message[:120]}")
        sequence += 1

    # Locate events.jsonl — for resumed sessions it's at a known path;
    # for new sessions we poll for the new directory.
    events_file: Path | None = None
    file_pos = 0

    if session_id:
        events_file = session_state_dir / session_id / "events.jsonl"
        # Seek to the end so we only stream *new* events from this interaction
        if events_file.exists():
            file_pos = events_file.stat().st_size
    else:
        # Wait up to 15s for a new session directory to appear
        for _ in range(30):
            await asyncio.sleep(0.5)
            current = {d.name for d in session_state_dir.iterdir() if d.is_dir()}
            new = current - existing_sessions
            if new:
                new_sid = new.pop()
                resolved_session_id = new_sid  # know the session_id as soon as dir appears
                events_file = session_state_dir / new_sid / "events.jsonl"
                break

    # Tail events.jsonl while copilot is running
    response_text = ""
    resolved_session_id = session_id or ""
    tool_names: dict[str, str] = {}  # toolCallId → toolName

    async def tail_once():
        nonlocal file_pos, response_text, resolved_session_id, sequence
        if not events_file or not events_file.exists():
            return
        with open(events_file) as f:
            f.seek(file_pos)
            new_content = f.read()
            file_pos = f.tell()
        for line in new_content.splitlines():
            line = line.strip()
            if not line:
                continue
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                continue
            # Extract session ID from session.start
            if event.get("type") == "session.start":
                resolved_session_id = event.get("data", {}).get("sessionId", resolved_session_id)
            # Extract final response text from assistant.message
            if event.get("type") == "assistant.message":
                content = event.get("data", {}).get("content", "")
                if content and not event.get("data", {}).get("toolRequests"):
                    response_text = content
            # Convert to chunk and post
            result = _event_to_chunk(event, tool_names)
            if result and chunk_url and send_ref:
                ctype, ccontent = result
                await _post_chunk(chunk_url, send_ref, resolved_session_id or session_id,
                                  agent_ref, namespace, sequence, ctype, ccontent)
                sequence += 1

    # Poll while process is running
    while proc.poll() is None:
        await tail_once()
        await asyncio.sleep(0.5)

    # One final tail pass to catch events written just before process exit
    await tail_once()

    # Close pipes immediately — do NOT read them. Copilot may have spawned children
    # (e.g. bash/sleep) that still hold the pipe write-end open; reading would block.
    returncode = proc.returncode
    if proc.stdout:
        proc.stdout.close()
    if proc.stderr:
        proc.stderr.close()

    # returncode -15 = SIGTERM (cancelled), -9 = SIGKILL
    if returncode in (-15, -9):
        cancelled_msg = "⛔ Request cancelled by user."
        if chunk_url and send_ref:
            await _post_chunk(chunk_url, send_ref, resolved_session_id or session_id,
                              agent_ref, namespace, sequence, "error", cancelled_msg)
        # Return cancelled message — _process_queue will POST it to the webhook
        return cancelled_msg, resolved_session_id or session_id or "unknown"

    if returncode != 0:
        if chunk_url and send_ref:
            await _post_chunk(chunk_url, send_ref, resolved_session_id or session_id,
                              agent_ref, namespace, sequence, "error",
                              f"copilot exited with code {returncode}")
        raise HTTPException(status_code=502, detail=f"copilot exited with code {returncode}")

    # Fallback: if we didn't get a response from events.jsonl, use a placeholder
    if not response_text:
        response_text = "No response captured"

    return response_text or "No response from copilot", resolved_session_id or "unknown"


@app.on_event("startup")
async def startup_event():
    asyncio.create_task(_process_queue())


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/asyncchat")
async def async_chat(req: AsyncChatRequest):
    """Fire-and-forget: enqueue message for background processing."""
    queue_id = str(uuid.uuid4())
    await _queue.put({
        "queue_id": queue_id,
        "message": req.message,
        "session_id": req.session_id,
        "send_ref": req.send_ref,
        "namespace": req.namespace,
        "agent_ref": req.agent_ref,
    })
    return {"queue_id": queue_id, "status": "queued"}


@app.delete("/cancel/{queue_id}")
async def cancel_queue_item(queue_id: str):
    """Kill the entire copilot process group for the given queue_id."""
    proc = _active_procs.get(queue_id)
    if proc is None:
        return {"status": "not_found", "queue_id": queue_id}
    if proc.poll() is not None:
        _active_procs.pop(queue_id, None)
        return {"status": "already_done", "queue_id": queue_id}
    try:
        pgid = os.getpgid(proc.pid)
        os.killpg(pgid, signal.SIGTERM)
        print(f"[cancel] sent SIGTERM to process group {pgid} for queue_id={queue_id}")
        # Give processes 3s to exit, then SIGKILL the group
        loop = asyncio.get_event_loop()
        try:
            await asyncio.wait_for(
                loop.run_in_executor(None, proc.wait), timeout=3.0
            )
        except asyncio.TimeoutError:
            os.killpg(pgid, signal.SIGKILL)
            print(f"[cancel] sent SIGKILL to process group {pgid}")
    except (ProcessLookupError, PermissionError):
        pass  # process already gone
    _active_procs.pop(queue_id, None)
    return {"status": "cancelled", "queue_id": queue_id}


@app.post("/chat", response_model=ChatResponse)
def chat(req: ChatRequest):
    # If the client passes a session_id, this is a follow-up turn.
    existing_session_id = req.session_id

    response_text, copilot_session_id = run_copilot(req.message, existing_session_id)

    # Use copilot's own session ID as the canonical ID.
    # Fall back to the client-supplied one, then a note that parsing failed.
    session_id = copilot_session_id or existing_session_id or "unknown"

    # If the client gave a different ID than what copilot returned (shouldn't
    # happen, but guard anyway), migrate the history file.
    if existing_session_id and copilot_session_id and existing_session_id != copilot_session_id:
        old_path = SESSIONS_DIR / f"{existing_session_id}.json"
        new_path = SESSIONS_DIR / f"{copilot_session_id}.json"
        if old_path.exists() and not new_path.exists():
            old_path.rename(new_path)

    history = load_session(session_id)
    history.append({"user": req.message, "assistant": response_text})
    save_session(session_id, history)

    return ChatResponse(response=response_text, session_id=session_id)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
