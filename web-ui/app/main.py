import asyncio
import json

from fastapi import FastAPI, Query
from fastapi.staticfiles import StaticFiles
from sse_starlette.sse import EventSourceResponse

from app.config import settings
from app import k8s_client

app = FastAPI(title="KubeCopilot Web UI")


@app.get("/health", include_in_schema=False)
async def health():
    return {"status": "ok"}


@app.get("/agents")
async def get_agents():
    return k8s_client.list_agents(settings.namespace)


@app.get("/sessions")
async def get_sessions(agent_ref: str = Query(...)):
    return k8s_client.list_sessions(agent_ref, settings.namespace)


@app.get("/running-sessions")
async def get_running_sessions(agent_ref: str = Query(...)):
    """Return sends that have no KubeCopilotResponse yet (in-progress)."""
    return k8s_client.list_running_sessions(agent_ref, settings.namespace)


@app.delete("/sessions")
async def delete_session(agent_ref: str = Query(...), session_id: str = Query(...)):
    """Delete all KubeCopilotSend and KubeCopilotResponse objects for the given session."""
    deleted = k8s_client.delete_session(session_id, agent_ref, settings.namespace)
    return {"deleted": deleted}


@app.get("/history")
async def get_history(agent_ref: str = Query(...), session_id: str = Query(...)):
    """Return full message history for a given session from KubeCopilotResponse objects."""
    return k8s_client.get_session_history(session_id, agent_ref, settings.namespace)


@app.get("/chunks/stream")
async def chunks_stream(
    agent_ref: str = Query(...),
    session_id: str = Query(default=""),
    send_ref: str = Query(default=""),
):
    """
    SSE stream of KubeCopilotChunk objects. Polls every 1s.
    Filter by send_ref (for real-time per-message activity) or session_id (for history).
    """
    async def generate():
        seen_sequences: set[int] = set()

        def fetch_chunks():
            if send_ref:
                return k8s_client.list_chunks_for_send(send_ref, agent_ref, settings.namespace)
            return k8s_client.list_chunks_for_session(session_id, agent_ref, settings.namespace)

        # Initial load
        try:
            for chunk in fetch_chunks():
                seen_sequences.add(chunk["sequence"])
                yield {"event": "chunk", "data": json.dumps(chunk)}
        except Exception as e:
            yield {"event": "error", "data": json.dumps({"message": str(e)})}
            return

        # Keep streaming new chunks
        while True:
            await asyncio.sleep(1)
            try:
                for chunk in fetch_chunks():
                    if chunk["sequence"] not in seen_sequences:
                        seen_sequences.add(chunk["sequence"])
                        yield {"event": "chunk", "data": json.dumps(chunk)}
            except Exception:
                pass

    return EventSourceResponse(generate())


@app.get("/chat/stream")
async def chat_stream(
    message: str = Query(...),
    agent_ref: str = Query(...),
    session_id: str = Query(default=""),
):
    async def generate():
        yield {"event": "status", "data": json.dumps({"message": "Submitting request to agent..."})}

        try:
            send_name = k8s_client.create_send(
                message, agent_ref, session_id or None, settings.namespace
            )
        except Exception as e:
            yield {"event": "error", "data": json.dumps({"message": str(e)})}
            return

        # Notify frontend immediately so it can start streaming activity chunks
        yield {"event": "started", "data": json.dumps({"send_ref": send_name})}
        yield {"event": "status", "data": json.dumps({"message": "Request queued — waiting for agent response..."})}

        elapsed = 0.0
        while True:
            await asyncio.sleep(settings.poll_interval)
            elapsed += settings.poll_interval

            try:
                resp = k8s_client.get_response_for_send(send_name, settings.namespace)
            except Exception as e:
                yield {"event": "error", "data": json.dumps({"message": str(e)})}
                return

            if resp is not None:
                spec = resp.get("spec", {})
                labels = resp["metadata"].get("labels", {})
                response_text = spec.get("response", "")
                new_session_id = spec.get("sessionID") or labels.get("kubecopilot.io/session-id", session_id)
                yield {
                    "event": "done",
                    "data": json.dumps({
                        "response": response_text,
                        "session_id": new_session_id,
                        "send_name": send_name,
                    }),
                }
                return
            else:
                yield {
                    "event": "heartbeat",
                    "data": json.dumps({"elapsed": int(elapsed), "phase": "Processing"}),
                }

        yield {"event": "error", "data": json.dumps({"message": "Unexpected loop exit"})}

    return EventSourceResponse(generate())


@app.post("/cancel")
async def cancel_send(send_ref: str = Query(...), agent_ref: str = Query(...)):
    """Create a KubeCopilotCancel to stop an in-flight send."""
    try:
        k8s_client.create_cancel(send_ref, agent_ref, settings.namespace)
        return {"status": "cancel_requested", "send_ref": send_ref}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


# Static files must be mounted last so API routes take precedence
app.mount("/", StaticFiles(directory="app/static", html=True), name="static")
