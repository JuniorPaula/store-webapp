# WebApp em Golang - Integração de Pagamento com Stripe
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)
![Git](https://img.shields.io/badge/git-%23F05033.svg?style=for-the-badge&logo=git&logoColor=white)
![MariaDB](https://img.shields.io/badge/MariaDB-003545?style=for-the-badge&logo=mariadb&logoColor=white)
![Socket.io](https://img.shields.io/badge/Socket.io-black?style=for-the-badge&logo=socket.io&badgeColor=010101)

O **WebApp em Golang** é uma aplicação web que oferece integração de pagamento com a Stripe. Esta aplicação tem como principal objetivo fornecer uma loja online onde os clientes podem realizar pagamentos de produtos utilizando cartões de crédito ou assinar planos. Além disso, o sistema inclui recursos de autenticação, comunicação via WebSocket e utiliza criptografia com bcrypt para garantir a segurança dos dados.

## Funcionalidades Principais

- **Integração de Pagamento com a Stripe**: Os clientes podem efetuar pagamentos de produtos utilizando cartões de crédito, proporcionando uma experiência de compra segura e conveniente.

- **Assinaturas de Planos**: Os usuários têm a opção de assinar planos, o que permite o acesso a recursos ou conteúdo exclusivo.

- **Sistema de Autenticação**: A aplicação inclui um sistema de autenticação para garantir que apenas usuários autorizados possam acessar determinadas áreas ou realizar ações.

- **Comunicação via WebSocket**: A comunicação em tempo real é suportada por meio de WebSockets, o que possibilita a interação em tempo real entre os usuários.

- **Segurança com Bcrypt**: As senhas dos usuários são armazenadas de forma segura utilizando o algoritmo de criptografia bcrypt, garantindo a proteção dos dados sensíveis.

## Tecnologias Utilizadas

- Golang: A aplicação backend é desenvolvida em Golang, uma linguagem conhecida pela sua eficiência e desempenho.

- Stripe: A Stripe é usada como serviço de pagamento para processar transações de cartão de crédito de forma segura.

- WebSocket: A comunicação em tempo real é implementada utilizando WebSockets.

- Bcrypt: A criptografia com bcrypt é empregada para proteger as senhas dos usuários.
- MariaDB: Como banco de dados principal.
- Docker: Para auxiliar no desenvolvimento


## Como Executar

1. Clone este repositório em sua máquina local.
2. Navegue até o diretório raiz do projeto.
3. Configure as variáveis de ambiente necessárias, como as chaves de API da Stripe e as configurações do banco de dados.
4. Soda CLI para executar as migrations: [Soda](https://gobuffalo.io/pt/documentation/database/soda/)
5. Execute a aplicação WEB Golang usando `go run ./cmd/web`.
6. Execute a aplicação SERVER Golang usando `go run ./cmd/api`.
7. Acesse a aplicação em seu navegador através do endereço correspondente.

## Uso
- Realize pagamentos de produtos utilizando cartões de crédito.
- Assine planos para acessar conteúdo exclusivo.
- Realize o login para acessar áreas restritas.
- Experimente a comunicação em tempo real via WebSocket.

## Contribuição

Contribuições para aprimorar e melhorar o projeto são bem-vindas! Sinta-se à vontade para enviar problemas (issues) ou solicitações de pull (pull requests).

## Licença

Este projeto está licenciado sob os termos da [Licença MIT](https://opensource.org/licenses/MIT).

## Contato

Se você tiver alguma dúvida ou sugestão, não hesite em entrar em contato através do email [luke.junnior@icloud.com](mailto:luke.junnior@icloud.com).