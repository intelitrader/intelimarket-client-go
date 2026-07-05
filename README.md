# O que é isso?
É um cliente para se conectar ao InteliMarket usando a linguagem Go. Ele é um go module.

Essa biblioteca é um binding que usa a lib da Intelitrader feita em C (p9mdi_client, binário incluso nesse repositório)

# Como compilar e rodar

- Instalar o Go (1.13+) se ainda não tiver feito
- Baixar esse repositório em ~/go/src (Se você mudou seu GOPATH você deve saber ajustar essas instruções de acordo)
- Configure a pasta de binários como fallback para carga de DLL/SO/DYLIB (essa lib é um wrapper de uma lib feita em C)
    - macOS: `export DYLD_FALLBACK_LIBRARY_PATH=~/go/src/intelimarket-client-go/internal/bin/`
    - Windows: TODO
    - Linux: `export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:~/go/src/intelimarket-client-go/internal/bin`
- Entrar na pasta do repositório e rodar `go install ./...` para compilar tudo e copiar para pasta `bin` do Go
- Pronto. Para testar se está tudo OK é só executar `~/go/bin/intelimarket-client-go-example`

# Exemplo de código
Veja em `cmd/intelimarket-client-go-example`
# teste
