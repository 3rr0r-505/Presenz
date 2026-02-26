# server/middleware/activity_middleware.py

from server.services.killswitch_service import KillSwitchService

class ActivityMiddleware:
    """
    ASGI Middleware to track request activity and update the KillSwitch timestamp.
    """

    def __init__(self, app, killswitch: KillSwitchService):
        self.app = app
        self.killswitch = killswitch

    async def __call__(self, scope, receive, send):
        # Only track HTTP requests
        if scope["type"] == "http":
            self.killswitch.update_activity()
        await self.app(scope, receive, send)