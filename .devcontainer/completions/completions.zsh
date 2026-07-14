# Shell completion for kubectl, kind and docker (zsh).
# Sourced from ~/.zshrc by .devcontainer/post-create.sh.

# Initialize the completion system in case oh-my-zsh hasn't.
autoload -Uz compinit && compinit -u
command -v kubectl &>/dev/null && source <(kubectl completion zsh)
command -v kind    &>/dev/null && source <(kind completion zsh)
command -v docker  &>/dev/null && source <(docker completion zsh 2>/dev/null) 2>/dev/null || true
alias k=kubectl
command -v kubectl &>/dev/null && compdef k=kubectl
