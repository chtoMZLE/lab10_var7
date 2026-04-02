"""
Верификация JWT-токенов, выпущенных Go-сервисом.

Алгоритм HS256, тот же разделяемый секрет что и в auth.go (JWTSecret).
"""
import jwt

# Должен совпадать с JWTSecret в go_service/auth.go
SECRET = "lab10-shared-secret-key-32bytes!!"
ALGORITHM = "HS256"


def verify_token(token: str) -> dict:
    """
    Декодирует и валидирует токен.
    Выбрасывает jwt.ExpiredSignatureError, jwt.InvalidSignatureError и др.
    при невалидном токене.
    """
    return jwt.decode(token, SECRET, algorithms=[ALGORITHM])


def get_username(token: str) -> str:
    """Возвращает username (поле 'username' или 'sub') из токена."""
    payload = verify_token(token)
    return payload.get("username") or payload.get("sub", "")


def is_token_valid(token: str) -> bool:
    """Возвращает True если токен валиден, False в любом ином случае."""
    try:
        verify_token(token)
        return True
    except jwt.PyJWTError:
        return False
