mkdir -p $HOME/.kube
echo $1 | base64 -d >$HOME/.kube/config