## RecursiveCall (Go)

### Description

Simple API that calls another instance of itself. Used to demonstrate Docker orchestration.

Related: [NodeJS version](https://github.com/davidbetz/recursivecall) 

### Details

This Go version creates a Docker image with a size of **4.26MB**.

The trick is to use create an image from scratch and make your Go binary statically linked.

### Docker Background

Docker "containers" aren't like virtual machines; you're not creating an general purpose environment with its own kernel. What you're doing is giving your project the minimum is needs to run.

#### Namespaces

Docker uses an existing Linux concept called **namespaces**. In fact, Docker is largely a tool that abstracts already existing Linux cleverness, including namespaces (and cgroups).

Namespaces in Linux serve a similar purpose as namespaces in C++/C#: they create a separate "area" where your names won't conflict. Types X.Y.Person and A.B.Person won't conflict. However, they aren't in different runtimes. They're isolated, but not entirely separate.

Linux processes and mounts (and other types of entities) can have different namespaces. Processes in one namespace don't conflict with processes in another. You can have a PID 1 in one namespace and PID 1 in a different namespace. The same goes for mounts. You can have /etc/hosts in one namespace and /etc/hosts in a different namespace. There's an isolation, but not an entire separate (e.g. they're using the same Kernel).

This is what Docker does with containers. There's just a lot of namespace magic to give the illusion of various "micro-machines". In reality, there are no "micro-machines"; everything is running in the same space, but with a simple label separating them. In fact, when you run `ps` in Linux, you see all the processes "in" each "container" (more accurately, you see all the process trees under the parent process tree).

> The term "container" and the preposition "in" lead to extreme confusion, but the terminology is pretty much baked into the industry at this point. There's nothing "in" a container, but something can be "in" an image.

Namespaces are clever and very helpful. If I were to write a plug-in model for an application, I'd create each plug-in in a different namespace then share an IPC namespace for communication. Supposedly, Google Chrome on Linux does something similar with namespaces for various add-ons. Namespaces give you an easy, built-in way to do jailing/sandboxing.

Because a mount namespace is isolated from another mount namespace, when files are required, they need to be in the correct namespace. In this case, isolation implies redundancy. Having /etc/hosts, for example, on the host doesn't help you, you need it in your namespace too. 

#### Images

Let's apply this information to Docker image creation.

When you're creating an image, you're mainly making careful decisions about what minimum files are required.

If you're trying to run a custom NodeJS API, you need to ask yourself: what's the the bare minimum required? The bare minimum is Node. Node, however, can't run by itself; it requires various shared libraries.

You could copy each of these libraries into the image or you could use something designed to satisfy most general share library requirements. The most common way to do this is to start your image with the [content from Alpine](https://github.com/gliderlabs/docker-alpine/blob/61c3181ad3127c5bedd098271ac05f49119c9915/versions/library-3.7/x86_64/Dockerfile).

> You could use Alpine, but I said *content from Alpine*. You want what it provides, if not Alpine as a whole.  You could literally take the `xz` file and run with it. That said, there's little reason not to base your work on Alpine.

We care less about a "base image" as we do the shared libraries within. Theoretically, if you're careful, you can create an image from `scratch`, dump in the Node binaries, then map each shared library from the host to the container.

There is, however, a simpler out-of-the-box model: **Go**.

With Go, you statically link your binary. No external libraries are required. No runtime is required. Your entire image is a single file: your binary.

In fact, your entire Docker image will have two layers: 1) copy the file, and 2) run the file.

That's it.

### Usage

To run directly:

    docker run -e GOPATH=/usr/src/app -v $PWD:/usr/src/app -w /usr/src/app golang go run app.go

To build:

    docker build . -t local/recursivecall-go

To get the full effect of this sample, run in a Swarm environment:

    docker stack deploy --compose-file docker-compose.yml --with-registry-auth rcgo

You need to add `g` to some form of name resolution (DNS or `/etc/hosts`). Once the call is inside Swarm, it will handle the rest of the name resolution itself.

## Kubernetes Deploy

Start:

    kubectl create -f k8s

Check:

    kubectl get service -l application.name=recursivecall-go

Stop:

    kubectl delete services,deployments -l application.name=recursivecall-go

Break service "c":

    kubectl delete deployment,service c
