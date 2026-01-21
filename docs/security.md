## Basic Auth

1. Client sends email and password over HTTPS to the login endpoint.
2. Server verifies the password and creates a session record with a new refresh token ID.
3. Server generates an access token (PASETO) signed with the current private key and includes a kid.
4. Server generates a refresh token (PASETO local) containing the refresh token ID and user ID.
5. Server returns both tokens to the client in a JSON response.
6. Client stores the access token in memory and the refresh token in secure storage.
7. Client sends the access token in the Authorization header for API requests.
8. Server verifies the access token signature using the public key identified by kid and checks expiry.
9. When the access token expires, the client sends the refresh token to the refresh endpoint.
10. Server verifies the refresh token, checks the session record is active, then issues a new access token.
11. Server rotates the refresh token by issuing a new refresh token, updating the session record, and revoking the old one.
12. Server rotates the signing key pair periodically and publishes the new public key in a key endpoint.
13. Client fetches the public key endpoint periodically.
14. Server rejects any token signed with a revoked key or using a missing kid.
