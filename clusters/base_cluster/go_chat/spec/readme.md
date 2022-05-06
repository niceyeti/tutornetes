# Go Chat

This is the go-based chat application. Users may load the base page and then submit and
view chat messages in a chat sidebar component.
The chat app need only serve the chat component and /chat endpoints, but for now serves the entire root page.

# High level goals
1) Code, code, code: golang review and design patterns
2) State replication in k8s
3) 100% test coverage, possibly integrate with git
The specific goal here is to implement SignalR in golang, as it (may) extend other realtime
use cases, such as sharing game state across multiplayer games.

# Development stages
1) Stage 1: no login, no user accounts, bare bones chat app. Users will
submit request containing chat content and a header "User: Bob"; their messages will
then be pushed down to all users of the app, reliably. Maybe deletion.
  * no security, login, formal users except for mock headers
  * vanilla js frontend and chat bar component plugin
  * no db
2) Stage 2: add database backend and polish requests (CQRS, even if bloated, will be good exercise)
  * Could add in-memory user+session model supported by db backend.
3) Stage 3: formal auth. This is the most bloated req, so leaving for last.
  * Keycloak et al
  * Find the simplest auth solution available: AD, LDAP, Keycloak, etc.
  * Minimize learning; I don't want to drag my feet learning third party auth components for this project. It is deserving of its own separate security project.

# Security
NOTE: Security is a gaping hole in this project and I have no intention of fixing it. This is not
a real chat application. In a real app, extensive anti-XSS measures would be required.

# HTTP Endpoints
Users can load all messages in the chat and submit their own; edit/delete functionality is not allowed, because reasons.

NOTE: although these will be implemented in SignalR, documenting here as HTTP for now.

1) /chat/[channel]/post: Users freely post messages here. They are pushed to all subscribers.
HTTP Verb: POST
Headers: "User: [Username]"
Future: derive user info for chat from jwt cookie; this depends on the gateway and auth, which are *future*.
2) /chat/[channel]/list: List all chat messages for users in channel.
HTTP Verb: GET
    NOTE: the messages need not be persisted, could well be in-memory for now, with acceptable loss
    on container restart.


   










