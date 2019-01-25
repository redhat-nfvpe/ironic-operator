# How to build

Ironic operator needs to be built using operator-sdk. In order to generate a
build for a given registry and repository, it can be done like:
```sh
operator-sdk build quay.io/yrobla/ironic-operator:v0.0.1
```

Then it can be pushed to the given registry:
```sh
docker push quay.io/yrobla/ironic-operator:v0.0.1
```

Also it needs to be considered that ironic-operator is storing some static file
content using packr. Every time that the `files` directory is modified,
resources need to be repacked. This can be achieved by going into `pkg/helpers`
directory, and executing the following calls:
```sh
packr -v
go build
```

After the resources are repacked, operator can be built and pushed, and will
contain the modified resources.
