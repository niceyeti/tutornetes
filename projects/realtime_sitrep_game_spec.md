# High level description

This is to be a dirt simple multiplayer game platform, whose general application
is realtime information sharing: video, games, etc.
* realtime chat
* realtime multiplayer games and data streaming/sharing (shared state)


# Primary learning goals

The goal here is to maximize time spent coding, playing with shared state methods in k8s clusters
and all of the problems therein, exercise a little postgres'ing, CQRS endpoints, and websocketeering.
A condition of this is to minimize time spent on things like auth, a worthy topic but also the most dependent on third party components. Similarly avoid other forms of bikeshedding and learning technologies or third party components.

# Unique contributions and learning
1) Shared state is hard
2) A repudiable, deterministic message system
3) Generic, turnkey universal information sharing system with "minimalist data hygiene"

# User stories
Users are authenticated and tracked on the platform, allowing them
to chat with others, share realtime information, and engage in games.

Platform administrators can view player stats (logins, session time, engagement metrics)
and deauthenticate users.

# Technical objectives and accomplishments
Though I'm not a gamer, game programming entails the full scope of programming skill
and seems more amenable to high-frequency, no-BS rapid development. A multiplayer
cloud-based platform steps this up even further since shared-state is a difficult
technical problem.

# Component-based design
Each responsibility listed above will be implemented by a separate 'team',
like microfrontends: 
1) chat functionality is a frontend component and backend stack
2) realtime game stack
3) user tracking stack: auth, metrics, etc.

# Infrastructure
The infrastructure will consist of:
1) BYO cloud:
  * K3S fo resiliency and development (k3d)
  * Standard (minimal) logging solution
  * Istio service mesh
2) Auth:
  * I'm thinking keycloak supports all requirements, which are:
    * Authentication
        * issue identities
        * track identity metrics (session, usage, etc.) using some service layer
            * user http headers
            * usage
            * login/logoff times
  * TODO: what out of the box identity solutions meet reqs?
3) Secure by default:
  * all persistent data encrypted at rest
  * no snooping
  * repudiable and self-deleting chat messages (like Signal)
  * manually driven kube-bench scanning
  * RBAC-modeled endpoints: 
    * all actions/data modeled as REST endpoints
    * endpoints secured by RBAC
    * equivalent to kubernetes' api-server RBAC
  * basic auditing and logging

# Schedule
1) chat engine: 
  * chat frontend and websockets
  * FUTURE: end to end encryption and repudiation
  * define REST endpoints:
    * POST
    * LIST
2) Auth layers: be able to create users, route based on access-controls, and redirect
to a login page when unauthorized.
  * auth solution
  * setup routes, security gateways
  * create/delete users
  * authorize users: routes, access rights, etc.
  * NOTE: this is the most dependent on third party components, and therefore least desirable. Minimize responsibilities and "let me go learn platform X..." behavior.
3) games

# Problems to solve
1) End to end encryption and repudiation requires a client app, inevitably in some
  native language. This is for key generation, encryption, and persistence, beyond
  what a browser can provide. This is why Signal, for example, is implemented as an
  Electron app. Specific properties:
    * end to end encryption
    * key persistence on client
    * message persistence on client
    * These properties, and therein trust boundaries, can be threat modeled to grasp and define their properties. 
  * Although complex, I believe this end-to-end scheme may be worthwhile precisely because of the client-side trust model. It has generic application to things like IOT/edge devices, machine learning models (which often can't fit in a browser), security/encryption (e.g. Signal).
    * Server-side responsibilities stay server side: authentication, L7/L4 streaming, platform health.
    * Client-side information and data remain client-side: repudiable information stays repudiable.
    * Recall we faced exactly these reqs at S**: edge monitors, device trust (writable airgapped devices), etc.
  * The client-side app headwinds are strong (Electron, etc) and coordinating multiple client-side services is a challenge. By contrast, browser-based pure-client web apps are simple, users simply trust the host company (lol).






