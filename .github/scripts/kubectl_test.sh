mkdir -p $HOME/.kube
echo "$KUBE_CONFIG" | base64 -d >$HOME/.kube/config
echo "FU"
echo "$(cat $HOME/.kube/config)"