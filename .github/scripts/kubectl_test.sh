mkdir -p $HOME/.kube
echo "$KUBE_CONFIG" | base64 -d
echo "FU"
echo "$(cat $HOME/.kube/config)"