# two_finance_auth_sdk

Dart SDK for the current `2finance-auth` HTTP API.

The recommended login flow for apps is OIDC Authorization Code + PKCE. The
password login route remains exposed only for compatibility with controlled
environments and should not be the default path for mobile, desktop, or browser
clients.

The SDK follows the routes currently exposed by `2finance-auth`:

- `POST /login` legacy compatibility
- `POST /refresh`
- `POST /logout`
- `POST /signup`
- `POST /phone/sms/request-code`
- `POST /phone/sms/login`
- `GET /pkce/login-redirect`
- `GET /pkce/callback`
- `GET /user-info`
- `POST /create-user`
- `POST /request-sms`
- `POST /verify-sms`

It intentionally does not expose routes that are not registered by the current
auth router, such as direct user security updates.

## Usage

```dart
import 'package:two_finance_auth_sdk/two_finance_auth_sdk.dart';

final auth = TwoFinanceAuthClient(
  baseUrl: Uri.parse('http://localhost:8080'),
  realm: '2Finance',
  clientId: '2finance-wallet-mobile',
  phoneClientId: '2finance-authenticator-phone',
);
```

## PKCE login

Use PKCE as the default authentication flow. Persist the refresh token only in
the operating system secure store and keep the access token in memory when
possible.

```dart
final login = await auth.loginPKCE();
// Redirect the user to login.authUrl.

final tokens = await auth.callbackPKCE(
  code: '<code from callback>',
  state: login.state,
  cookies: login.cookies,
);

final userInfo = await auth.getUserInfo(tokens.accessToken);
```

## Refresh token rotation

`refreshToken` returns a new token set. Treat the returned `refresh_token` as the
only valid refresh token and replace the stored value atomically.

```dart
final rotated = await auth.refreshToken(storedRefreshToken);

await secureStorage.write(
  key: '2finance.refresh_token',
  value: rotated.refreshToken,
);
```

If refresh fails with 401 or 403, clear local tokens and send the user through
PKCE again.

## Logout

Call `logout` with the current refresh token to revoke the server-side session,
then clear local sensitive state even if revocation fails.

```dart
final storedRefreshToken = await secureStorage.read(
  key: '2finance.refresh_token',
);
if (storedRefreshToken != null) {
  await auth.logout(storedRefreshToken);
}
await secureStorage.delete(key: '2finance.refresh_token');
accessTokenCache.value = null;
```

## Legacy password login

`login(LoginInput)` is deprecated in the SDK API. Keep it for tests,
development-only tools, or controlled migrations where PKCE is not available.
Do not collect or embed user passwords in production apps when PKCE can be used.
