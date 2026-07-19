# Authentication and Authorization

## Papel do repo

`2finance-auth-dart-sdk` deve ser o cliente Dart/Flutter para `2finance-auth`, priorizando OIDC Authorization Code + PKCE.

## Avaliacao atual

- O README lista rotas atuais de login, refresh, logout, signup, phone, user-info e PKCE.
- O fluxo PKCE ja existe no SDK, mas o login com usuario/senha deve ser tratado como legado ou restrito a ambientes controlados.
- `TwoFinanceAuthClient.login` e `LoginInput` ficam disponiveis apenas para compatibilidade e estao marcados como deprecated na API Dart.

## Padrao seguro obrigatorio

- Usar PKCE como fluxo padrao para mobile/desktop/browser.
- Nao embutir `client_secret` no app.
- Guardar refresh token apenas em storage seguro do sistema operacional.
- Manter access token em memoria quando possivel.
- Nao enviar `X-User-ID` ou `X-Tenant-ID` como autoridade de identidade; derivar esses valores das claims validadas pelo backend.
- Mascarar tokens em exceptions, traces e logs de debug.

## Login por senha legado

`POST /login` continua listado porque a rota existe no `2finance-auth`, mas nao deve ser o caminho recomendado para apps. Use apenas em ferramentas internas, testes, migracoes controladas ou ambientes onde PKCE ainda nao esteja disponivel.

Para mobile, desktop e browser, a tela de login deve redirecionar para o provedor OIDC via `loginPKCE()` e trocar o callback com `callbackPKCE()`.

## Refresh token rotation

O refresh token deve ser tratado como segredo de longa duracao e guardado apenas em storage seguro do sistema operacional, por exemplo Keychain, Keystore ou equivalente Flutter.

Fluxo recomendado:

1. Ler o refresh token do storage seguro apenas no momento do refresh.
2. Chamar `refreshToken(storedRefreshToken)`.
3. Substituir o refresh token armazenado pelo `refresh_token` retornado na mesma operacao logica.
4. Atualizar o access token em memoria.
5. Se o refresh falhar com 401 ou 403, apagar tokens locais e reiniciar o fluxo PKCE.

Exemplo:

```dart
final storedRefreshToken = await secureStorage.read(
  key: '2finance.refresh_token',
);

if (storedRefreshToken == null) {
  // Inicie loginPKCE().
  return;
}

try {
  final rotated = await auth.refreshToken(storedRefreshToken);
  accessTokenCache.value = rotated.accessToken;
  await secureStorage.write(
    key: '2finance.refresh_token',
    value: rotated.refreshToken,
  );
} on AuthSdkException catch (error) {
  if (error.statusCode == 401 || error.statusCode == 403) {
    accessTokenCache.value = null;
    await secureStorage.delete(key: '2finance.refresh_token');
    // Inicie loginPKCE() novamente.
  }
  rethrow;
}
```

## Logout

O SDK expoe `logout(refreshToken)` para revogar a sessao server-side no
`2finance-auth`. O app continua responsavel por limpar estado local mesmo quando
a chamada de revogacao falhar:

1. Ler o refresh token do storage seguro.
2. Chamar `logout(storedRefreshToken)` quando existir refresh token.
3. Remover o access token da memoria.
4. Apagar o refresh token do storage seguro.
5. Descartar cookies temporarios do fluxo PKCE mantidos pelo app.
6. Limpar estado de usuario em stores locais.

Erros de logout passam por `AuthSdkException` e redigem `refresh_token`,
`access_token`, `id_token`, `password` e bearer tokens antes de aparecerem em
logs/excecoes.

## Exemplo de integracao com `2finance-wallet-mobile`

O `2finance-wallet-mobile` deve iniciar o login pelo SDK, abrir a URL de autorizacao no navegador/sistema, receber o deep link de callback e entao chamar `callbackPKCE`.

```dart
class AuthCoordinator {
  AuthCoordinator({
    required this.auth,
    required this.secureStorage,
  });

  final TwoFinanceAuthClient auth;
  final SecureStorage secureStorage;

  PKCELoginResponse? _pendingLogin;
  String? accessToken;

  Future<Uri> startLogin() async {
    final login = await auth.loginPKCE();
    _pendingLogin = login;
    return login.authUrl;
  }

  Future<void> completeLogin(Uri callbackUri) async {
    final pending = _pendingLogin;
    if (pending == null) {
      throw StateError('PKCE login was not started');
    }

    final code = callbackUri.queryParameters['code'];
    final state = callbackUri.queryParameters['state'];
    if (code == null || state == null) {
      throw StateError('PKCE callback missing code or state');
    }

    final tokens = await auth.callbackPKCE(
      code: code,
      state: state,
      cookies: pending.cookies,
    );

    accessToken = tokens.accessToken;
    await secureStorage.write(
      key: '2finance.refresh_token',
      value: tokens.refreshToken,
    );
    _pendingLogin = null;
  }

  Future<void> logout() async {
    final storedRefreshToken = await secureStorage.read(
      key: '2finance.refresh_token',
    );
    if (storedRefreshToken != null) {
      await auth.logout(storedRefreshToken);
    }
    accessToken = null;
    _pendingLogin = null;
    await secureStorage.delete(key: '2finance.refresh_token');
  }
}
```

`SecureStorage` acima representa a abstracao ja usada pelo app para storage seguro. O SDK nao deve escolher o plugin de storage nem persistir tokens diretamente.

## Proximos passos

Fechados neste SDK:

1. Login com senha marcado como compatibilidade, nao como caminho recomendado.
2. Guia de refresh token rotation e logout adicionado.
3. Exemplo de integracao com `2finance-wallet-mobile` usando PKCE adicionado.
4. Excecoes do SDK mascaram credenciais e tokens conhecidos antes de expor o corpo do erro.
5. Logout server-side adicionado via `TwoFinanceAuthClient.logout(refreshToken)`.

Restante fora deste SDK:

1. Garantir que `2finance-wallet-mobile` use storage seguro real e nao persista access token em disco.
