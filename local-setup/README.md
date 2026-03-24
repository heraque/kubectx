# Local setup

Artifacts para reinstalar seu fluxo local sem depender de Homebrew.

Conteudo:

- `bin/darwin-arm64/kubectx`
- `bin/linux-amd64/kubectx`
- `bin/linux-arm64/kubectx`
- `bin/kubectl`: wrapper que agrega todos os kubeconfigs validos de `~/.kube`
- `install.sh`: instala o binario correto para a plataforma, o wrapper `kubectl`, cria `alias ctx="kubectx"` e ajusta `~/.zprofile` no macOS ou `.zshrc`/`.bashrc` no Linux

Uso padrao:

```sh
cd local-setup
./install.sh
```

Exemplos:

```sh
# macOS: instala em ~/.local/bin e garante PATH
./install.sh

# Linux: por padrao instala em /usr/local/bin
sudo ./install.sh

# Override manual
./install.sh --os linux --arch amd64 --target-dir "$HOME/bin"
```

Observacoes:

- O wrapper `kubectl` delega para um `kubectl` real ja instalado na maquina.
- O `kubectx` e empacotado por plataforma/arquitetura; binario de macOS nao roda em Linux.
