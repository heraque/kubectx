# Local setup

Artifacts para reinstalar seu fluxo local sem depender de Homebrew.

Conteudo:

- `bin/kubectx`: binario buildado desta branch para macOS local.
- `bin/kubectl`: wrapper que agrega todos os kubeconfigs validos de `~/.kube`.
- `install.sh`: instala ambos em `~/.local/bin` e garante `PATH`.

Uso:

```sh
cd local-setup
./install.sh
exec zsh -l
```

Observacoes:

- O wrapper `kubectl` delega para um `kubectl` real ja instalado na maquina.
- O binario `kubectx` aqui e um build local para a arquitetura atual.
