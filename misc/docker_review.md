## Cmd Line

Recall: Docker daemon and containers run as root by default. Rootless docker is available as of
version 20.10.

* docker build -t some_repo/some_image .
* docker build -t some_repo/some_image -f ../Dockerfile
    * --rm: remove intermediate build containers
    * --output type=tar,dest=out.tar: save image locally to out.tar file
* docker diff: view changes to a containers file system
* docker attach/exec: attach to a running container or run commands inside
* docker cp ./some_file CONTAINER:/some_file
* docker cp CONTAINER:/var/logs/app.log - | tar x -O | grep "ERROR"
    * Copies to stdout and pipes to other commands
* docker save -o fedora-all.tar fedora
* docker load img.tar : load an image from a tar
* docker ps -a : show all running containers
* docker port [container] : show all ports used by container
* docker image prune : prune unused images
* docker system prune : prune images, network, and other objects
* docker logs [container]
* docker commit [container] : create a new image from running container
* docker rm : remove a stopped container
* docker rmi : remove an image
* docker history --no-trunc : show the build process of an image; useful for reverse engineering a Dockerfile from an image.
* docker run -v [HOSTDIR]:[CONTAINER_DIR] some_image

## Docker Files

Of note:
* every line in a docker file outputs a new layer in the image, even if squashed at runtime

FROM: alpine, scratch, etch.
    * Use minimal distro, e.g. alpine
    * Or use scratch for images built from raw executables
    * Also use scratch for bare VOLUME images

CMD: without brackets runs command as a child process of the shell:
```
    # Same as: /bin/sh -c 'echo $PATH'
    CMD echo $PATH
```
Providing brackets indicates the exe will be run directly. The benefit is that SIGINT and similar signals will be caught.

## Patterns

1) Use multistage builds to reduce image size, or even scratch if possible
2) Data containers: docker supports a VOLUME instruction whereby an image may consist entirely of data. This supports some interesting use-cases, such as machine learning on secured data:
* Build and distribute datasets as encrypted containers
* Copy them into Pods using init-containers and sidecars, and Secrets for threat model.
* Thereby, one could build a set of ML/analysis services with only read-only capabilities on the data, which remains encrypted throughout. In transit, and at rest, though the precise way of stating this model is simply that data would be minimally unencrypted and only within specific contexts.

## Other

Then of course the canonical cheatsheet:

![Docker cheatsheet](./dockercheat.png)



