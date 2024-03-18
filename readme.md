
#### What is here?

I'm a big fan of the pixel madness called ragnarok online especially old good pre-renewal and I ❤️ Kubernetes. In this repo you can find some code to deploy rathena server to Kubernetes local for solo gaming as well as online to play with friends.

#### Runnig local server

I use [rancher-desktop](https://github.com/rancher-sandbox/rancher-desktop/releases). deployment is simple
```
kustomize 
kubectl apply -k deployments/local/. 
#work in progress. you need to forward ports and other things
```

#### Running public server

Why else to write the code if don't run it? [xxbbxb.github.io/open-ro/](https://xxbbxb.github.io/open-ro/) is real ragnarok server running in Kubernetes using code from this repo without additional modifications. Run you own server if you want or use Open-RO

#### Have a questions?

Raise an issue or text me on [telegram]https://t.me/openrohelp)