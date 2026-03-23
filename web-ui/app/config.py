from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    namespace: str = "ghcopilot-agent"
    default_agent_ref: str = ""
    poll_interval: float = 2.0
    command_timeout: int = 1800

    class Config:
        env_prefix = "WEBUI_"


settings = Settings()
