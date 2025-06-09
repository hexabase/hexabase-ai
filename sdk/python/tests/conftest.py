"""Pytest configuration and shared fixtures."""

import pytest
import asyncio
from typing import Generator


@pytest.fixture(scope="session")
def event_loop() -> Generator:
    """Create an instance of the default event loop for the test session."""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
def mock_api_response():
    """Factory for creating mock API responses."""
    def _make_response(status_code=200, json_data=None, headers=None):
        class MockResponse:
            def __init__(self):
                self.status_code = status_code
                self.headers = headers or {}
                self._json_data = json_data or {}
                
            def json(self):
                return self._json_data
                
            def raise_for_status(self):
                if self.status_code >= 400:
                    raise Exception(f"HTTP {self.status_code}")
                    
        return MockResponse()
    
    return _make_response