# 2Finance SDK Client

Coleção de SDKs oficiais para integrar aplicações à plataforma 2Finance. O
repositório mantém contratos e clientes em C++, Dart, Go, Java, JavaScript, PHP,
Python e TypeScript, com testes compartilhados de estrutura e comportamento.

## Quando usar

Use o SDK da linguagem da aplicação para consumir APIs e contratos 2Finance em
vez de duplicar modelos, serialização, autenticação ou tratamento de erros. Este
é um pacote de bibliotecas e não inicia um servidor ou publica uma porta.

## Estrutura

| Diretório | Plataforma |
| --- | --- |
| `cpp/` | C++ |
| `dart/` | Dart/Flutter e módulos auth/blockchain |
| `go/` | Go |
| `java/` | Java/Maven |
| `javascript/` | JavaScript |
| `php/` | PHP/Composer |
| `python/` | Python |
| `typescript/` | TypeScript |
| `contracts/` | Contratos compartilhados |

Cada diretório possui um README com instalação e exemplos específicos.

## Build e testes

Execute toda a matriz:

```bash
make test
```

Ou valide apenas uma linguagem:

```bash
make go-test
make java-test
make javascript-test
make python-test
make typescript-test
make dart-test
make cpp-test
make php-test
```

Use `make sdk-structure-test` para validar o layout e `make examples-test` para
os exemplos. Toolchains de todas as linguagens são necessários apenas para a
matriz completa.

## Uso local

Consulte o README da linguagem escolhida e aponte o cliente para os endpoints do
ambiente integrado. Os principais são Exchange
<https://exchange.localhost:38443/>, Explorer
<https://explorer.localhost:38443/> e, dentro do cluster, Network em
`svc-2finance-network.2finance-local.svc.cluster.local:9095`.

## Links

- Repositório: <https://github.com/2Finance-Labs/2finance-sdk-client>
- Índice do workspace e todos os endpoints: [`../README.md`](../README.md)
- Contratos: [`contracts/README.md`](contracts/README.md)

Não inclua chaves privadas, tokens ou credenciais em exemplos e fixtures.
