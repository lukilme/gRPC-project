# Microservices – Ambiente de Desenvolvimento

Este repositório contém um conjunto de microserviços escritos em Go (Order, Payment e Shipping), além de um **Makefile** para facilitar a execução local em ambiente de desenvolvimento.

A proposta é simples: cada serviço roda de forma independente, conecta-se a um banco MySQL e se comunica via HTTP/gRPC conforme necessário.

---

## Pré-requisitos

Antes de usar os comandos do Makefile, garanta que você tenha:

* Go instalado (versão compatível com o projeto)
* Docker e Docker Compose
* MySQL rodando em container com o nome:

  ```
  microservices-mysql
  ```
* Usuário do banco:

  * **user:** root
  * **password:** password

---

## Estrutura dos serviços

Cada microserviço segue a mesma organização básica:

```
microservices/
├── order/
│   └── cmd/main.go
├── payment/
│   └── cmd/main.go
└── shipping/
    └── cmd/main.go
```

Cada serviço:

* Possui seu próprio banco de dados
* Define suas variáveis de ambiente no momento da execução
* Roda localmente com `go run`

---

## Comandos do Makefile

### Subir o Payment Service

```bash
make run-payment
```

Configuração:

* Porta: **3001**
* Banco: **payment**
* Driver: MySQL
* Ambiente: development

---

### Subir o Order Service

```bash
make run-order
```

Configuração:

* Porta: **3000**
* Banco: **order**
* Driver: MySQL
* Integração com Payment Service em `localhost:3001`
* Ambiente: development

---

### Subir o Shipping Service

```bash
make run-shipping
```

Configuração:

* Porta: **3002**
* Banco: **shipping**
* Driver: MySQL
* Ambiente: development

---

### Testar conexão com o MySQL

```bash
make test-db
```

Executa um comando simples dentro do container MySQL para listar os bancos disponíveis.

Útil para validar se o container está ativo e acessível.

---

### Criar bancos de dados

```bash
make setup
```

Cria o banco **payment** caso ele ainda não exista.

> Observação: outros bancos (order, shipping) podem ser criados manualmente ou adicionados a este comando futuramente.

---

## Observações importantes

* Todos os serviços assumem que o MySQL está acessível em `127.0.0.1:3306`
* As credenciais estão fixas no Makefile e **não devem ser usadas em produção**
* Este setup é voltado exclusivamente para desenvolvimento local

---

## Próximos passos sugeridos

* Centralizar configuração via `.env`
* Criar um `docker-compose.yml` para subir tudo com um comando
* Adicionar targets de migração de banco
* Padronizar logs e health checks

---

Ambiente simples, explícito e previsível — do jeito certo para desenvolver e depurar microserviços.

---

## Visão geral da arquitetura

O projeto adota uma **arquitetura de microserviços clássica**, com separação clara de responsabilidades, comunicação via **gRPC** e persistência em **MySQL**. Cada serviço é isolado, possui seu próprio banco de dados e segue princípios de **Clean Architecture / Ports and Adapters**.

Os principais componentes são:

* **Order Service**: orquestra pedidos e integra integrações externas
* **Payment Service**: responsável por processamento e persistência de pagamentos
* **Shipping Service**: responsável por informações e fluxo de envio
* **microservices-proto**: definição e geração centralizada dos contratos gRPC

---

## Organização do repositório

```
.
├── docker-compose.yml     # Sobe infraestrutura (MySQL, etc.)
├── init.sql               # Script inicial de criação de bancos/tabelas
├── main.sh                # Script auxiliar de inicialização
├── Makefile               # Comandos de execução e utilidades
├── microservices/         # Implementação dos serviços
├── microservices-proto/   # Contratos gRPC e código gerado
└── README.md
```

---

## Padrão interno dos microserviços

Todos os serviços seguem a mesma estrutura conceitual:

```
cmd/            # Ponto de entrada da aplicação
config/         # Leitura e validação de variáveis de ambiente
internal/
  adapters/     # Implementações concretas (DB, gRPC, HTTP, etc.)
  application/  # Regras de negócio
    core/       # Domínio e casos de uso
    ports/      # Interfaces (contratos)
```

Esse padrão facilita:

* Testes unitários
* Substituição de dependências
* Evolução independente dos serviços

---

## Comunicação entre serviços

A comunicação entre os microserviços ocorre via **gRPC**, com contratos definidos em:

```
microservices-proto/proto/*.proto
```

O código Go gerado fica em:

```
microservices-proto/golang/
```

Esse módulo é importado pelos serviços consumidores, garantindo:

* Tipagem forte
* Versionamento explícito
* Contratos compartilhados

---

## Geração de Protos

Para gerar ou atualizar os arquivos `.pb.go`:

```bash
cd microservices-proto
./generator_proto.sh
```

Regras importantes:

* Nunca editar arquivos `.pb.go` manualmente
* Alterações devem ser feitas apenas nos `.proto`

---

## Infraestrutura com Docker

O arquivo `docker-compose.yml` é responsável por subir a infraestrutura local, incluindo o MySQL.

Inicialização típica:

```bash
docker compose up -d
```

O script `init.sql` cria os bancos necessários (`order`, `payment`, `shipping`).

---

## Fluxo recomendado de desenvolvimento

1. Subir a infraestrutura:

   ```bash
   docker compose up -d
   ```

2. Criar os bancos:

   ```bash
   make setup
   ```

3. Gerar contratos gRPC (se necessário)

4. Subir os serviços individualmente:

   ```bash
   make run-payment
   make run-order
   make run-shipping
   ```

---

## Observações finais

* Cada serviço possui seu próprio `go.mod`
* Não há dependência direta entre bases de dados
* Configurações são feitas exclusivamente via variáveis de ambiente

Projeto estruturado para crescer sem perder controle — simples, explícito e sustentável.

### Estrutura de diretórios
```sh
├── client
│   ├── domain
│   │   └── order.go
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   └── tui
│   │       ├── form.go
│   │       ├── grpc
│   │       │   └── client.go
│   │       ├── model.go
│   │       ├── update.go
│   │       └── view.go
│   └── main
│       └── main.go
├── curl
│   ├── order.proto
│   ├── order_test.proto
│   └── payment.proto
├── docker-compose.yml
├── init.sql
├── main.sh
├── Makefile
├── microservices
│   ├── order
│   │   ├── cmd
│   │   │   └── main.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── internal
│   │   │   ├── adapters
│   │   │   │   ├── db
│   │   │   │   │   └── mysql.go
│   │   │   │   ├── grpc
│   │   │   │   │   └── server.go
│   │   │   │   └── payment
│   │   │   │       └── payment.go
│   │   │   └── application
│   │   │       ├── core
│   │   │       │   ├── api
│   │   │       │   │   └── api.go
│   │   │       │   └── domain
│   │   │       │       └── order.go
│   │   │       └── ports
│   │   │           └── payment.go
│   │   ├── proto
│   │   │   ├── order.proto
│   │   │   ├── payment.proto
│   │   │   └── shipping.proto
│   │   ├── proto copy
│   │   │   └── payment.proto
│   │   └── test
│   │       └── main_test.go
│   ├── payment
│   │   ├── cmd
│   │   │   └── main.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── DB_DRIVER=mysql
│   │   ├── deployment.yaml
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── internal
│   │       ├── adapters
│   │       │   ├── db
│   │       │   │   └── db.go
│   │       │   └── grpc
│   │       │       ├── grpc.go
│   │       │       └── server.go
│   │       ├── application
│   │       │   └── core
│   │       │       ├── api
│   │       │       │   └── api.go
│   │       │       └── domain
│   │       │           └── payment.go
│   │       └── ports
│   │           ├── api.go
│   │           └── db.go
│   └── shipping
│       ├── cmd
│       │   └── main.go
│       ├── config
│       │   └── config.go
│       ├── go.mod
│       ├── go.sum
│       └── internal
│           ├── adapters
│           │   ├── db
│           │   │   └── mysql.go
│           │   └── grpc
│           │       └── server.go
│           ├── application
│           │   ├── core
│           │   │   ├── api
│           │   │   │   └── api.go
│           │   │   └── domain
│           │   │       └── shipping.go
│           │   └── ports
│           ├── domain
│           └── ports
├── microservices-proto
│   ├── generator_proto.sh
│   ├── golang
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── order
│   │   │   ├── order_grpc.pb.go
│   │   │   └── order.pb.go
│   │   ├── payment
│   │   │   ├── payment_grpc.pb.go
│   │   │   └── payment.pb.go
│   │   └── shipping
│   │       ├── shipping_grpc.pb.go
│   │       └── shipping.pb.go
│   ├── order_grpc.pb.go
│   ├── order.pb.go
│   ├── payment_grpc.pb.go
│   ├── payment.pb.go
│   ├── proto
│   │   ├── order.proto
│   │   ├── payment.proto
│   │   └── shipping.proto
│   ├── shipping_grpc.pb.go
│   └── shipping.pb.go
├── README.md
├── run.sh
└── send.sh
```