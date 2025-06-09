"""Exception classes for Hexabase AI SDK."""


class HexabaseError(Exception):
    """Base exception for all Hexabase AI SDK errors."""
    
    def __init__(self, message: str, code: str = None, details: dict = None):
        super().__init__(message)
        self.code = code
        self.details = details or {}


class AuthenticationError(HexabaseError):
    """Raised when authentication fails."""
    pass


class TokenExpiredError(AuthenticationError):
    """Raised when access token has expired."""
    pass


class FunctionNotFoundError(HexabaseError):
    """Raised when a function is not found."""
    pass


class FunctionExecutionError(HexabaseError):
    """Raised when function execution fails."""
    pass


class ValidationError(HexabaseError):
    """Raised when input validation fails."""
    pass


class NetworkError(HexabaseError):
    """Raised when network operations fail."""
    pass


class RateLimitError(HexabaseError):
    """Raised when API rate limit is exceeded."""
    pass


class TimeoutError(HexabaseError):
    """Raised when an operation times out."""
    pass