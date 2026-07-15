concorrência
============

Olá, sou o Pedro, dono deste repositório, e vou confessar uma coisa: eu tenho
certa implicância com o Virtual DOM e com todas as frameworks baseadas nele,
mas reconheço a utilidade delas. Recentemente, porém, tive contato com o Lit
(lit.dev), uma ferramenta que permite utilizar o Shadow DOM de uma forma um
pouco mais facilitada, além de gerenciar melhor as modificações recebidas via
WebSocket, agrupando-as em lotes para atualizar em tempo real.

Fiz este projeto baseado justamente nisso. Utilizei a concorrência do Go no
backend para me auxiliar nesse teste. A ideia é que eu possa rodar algo no
estilo do Siege, no qual posso colocar grupos de usuários fazendo débito e
crédito diretamente em uma conta compartilhada, como se estivessem depositando e
retirando dinheiro, definindo a quantidade de usuários e de requisições, todos
adicionando ou retirando o valor 1 da conta.

Eu queria ver isso acontecendo em tempo real e entender como o Lit atualiza o
dado na tela: quanto tempo demora, se é possível conectar em computadores
diferentes, com navegadores diferentes, compartilhando a mesma tela, para ver
se a atualização ocorre em tempo real e se o sistema trava em algum momento.

Portanto, este repositório serve apenas para estudo, para verificar se tudo
está funcionando corretamente, se o sistema resiste sob estresse e assim por
diante. Caso queiram saber mais sobre o projeto, fiquem à vontade. Abaixo está
a documentação :)

---

Laboratório didático full-stack para estudar padrões de concorrência em Go e
reatividade com Shadow DOM via Lit. Uma única conta bancária compartilhada é
operada por múltiplos usuários simultâneos via HTTP, com propagação de saldo em
tempo real para um frontend construído com Web Components. O sistema elimina
condições de corrida e atualizações perdidas sob carga arbitrária.

Características
---------------
- Backend em Go com transações SQLite em modo WAL protegidas por `sync.Mutex`
  para operações atômicas de crédito (+1) e débito (-1)
- Broadcast em tempo real de alterações de saldo via WebSocket para todos os
  clientes conectados
- Orquestrador de teste de carga configurável com grupos de goroutines
  concorrentes
- Duas telas Lit com Shadow DOM: configuração de teste de carga (Tela 1) e
  visualização de saldo ao vivo (Tela 2)

Instalação
------------

Requisitos: Go 1.22+ (o `go.mod` aponta para 1.26.5).

    git clone https://github.com/pedro/concorrencia.git
    cd concorrencia
    go build .

Uso
---

    go run .

Abra `http://localhost:8080` em uma aba para configurar o teste de carga (Tela 1)
e `http://localhost:8080/saldo.html` em outra aba para acompanhar o saldo em
tempo real (Tela 2).

Também é possível usar o Makefile:

    make all      # compila e inicia o servidor em background
    make stop     # encerra o servidor
    make re       # recria do zero

Endpoints da API
----------------

| Método | Endpoint            | Descrição                           |
|--------|---------------------|-------------------------------------|
| POST   | `/credit`           | Credita +1 na conta compartilhada   |
| POST   | `/debit`            | Debita -1 da conta compartilhada    |
| GET    | `/balance`          | Retorna o saldo atual em JSON       |
| POST   | `/load-test/start`  | Inicia orquestrador de teste de carga |
| GET    | `/ws`               | WebSocket para atualizações de saldo |

Exemplo de requisição de teste de carga:

    curl -X POST http://localhost:8080/load-test/start \
      -H "Content-Type: application/json" \
      -d '[{"users":10,"requests":50,"type":"credit"},
           {"users":5,"requests":20,"type":"debit"}]'

Configuração
------------

Variáveis de ambiente:

- `PORT` — porta do servidor HTTP (padrão: `8080`)
- `DB_PATH` — caminho do banco SQLite (padrão: `concorrencia.db`)

Arquitetura
-----------

- `internal/db` — abre SQLite com WAL mode, executa migrações e garante a
  existência da conta singleton
- `internal/account` — lógica de domínio com mutex e transações para evitar
  lost updates
- `internal/ws` — hub WebSocket que transmite o saldo para todos os clientes
  conectados
- `internal/loadtest` — orquestrador que dispara goroutines virtuais contra os
  endpoints da API
- `frontend/app` — componentes Lit (`load-test-config`, `balance-view`,
  `nav-menu`) com Shadow DOM e estética brutalista

Testes
------

    go test -race ./...

A suíte cobre:
- Operações de crédito e débito
- Concorrência mista (100 créditos + 100 débitos simultâneos)
- Broadcast WebSocket após mutação
- Relatório do orquestrador com múltiplos grupos

Licença
-------
MIT
