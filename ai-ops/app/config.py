import os
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    HKS_INTERNAL_API_URL: str = os.getenv("HKS_INTERNAL_API_URL", "http://api-service.hexabase.svc.cluster.local")

settings = Settings() 