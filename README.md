# docker-utils

External-to-docker utilities to provide a more complete experience


# Commands

## dockertarsum

Like the system sum utilites (md5sum, sha1sum, sha256sum, etc), this is a
command line tool to get the fixed time checksum of docker image layers.

### Installing

	go get github.com/vbatts/docker-utils/cmd/dockertarsum

### Usage

The tool is very similar behavior to the system sum utilites like sha1sum, but
due to the nature of image layers being a tar and json, and that is usually
included in a larger tar archive, some of the same explicit path notations do
not work the same.

For a basic example, saving an image, and getting the tarsum for its layers:

	$ docker save -o busybox.tar busybox
	$ dockertarsum ./busybox.tar > busybox.tar.sums
	$ cat busybox.tar.sums
	tarsum+sha256:7b0ade22d5bba35d1e88389c005376f441e7d83bf5f363f2d7c70be9286163aa  ./busybox.tar:120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16
	tarsum+sha256:caf47d0e2cbedb826aa7205464b807698101b2e43b383b58ad9b28fe3a31499e  ./busybox.tar:42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229
	tarsum+sha256:89c1b4a3aeb8d03f77f6ebfe8e8c894b5939102fbdad110626ff05bc94a94d56  ./busybox.tar:511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158
	tarsum+sha256:aa441c90dc4e77d74b715921ddd63d2e8cfe22d5aae753a550b7e87f9010fee4  ./busybox.tar:a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721

Checking the sums saved, against tar achive:

	$ dockertarsum -c ./busybox.tar.sums ./busybox.tar
	120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16:  OK
	42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229:  OK
	511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158:  OK
	a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721:  OK
	$ echo $?
	0

Checking the sums saved, against stdin image straight from docker:

	$ docker save busybox | dockertarsum -c ./busybox.tar.sums
	120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16:  OK
	42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229:  OK
	511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158:  OK
	a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721:  OK
	$ echo $?
	0

A failed checked:

	$ docker save my-app | dockertarsum -c ./my-app.tar.sums
	0c752394b855e8f15d2dc1fba6f10f4386ff6c0ab6fc6a253285bcfbfdd214f5:  OK
	17ed77acc78f8e49a8efdfa209a7ff16e774622135c661b05f11fea0f57a41e3:  FAILED
	34e94e67e63a0f079d9336b3c2a52e814d138e5b3f1f614a0cfe273814ed7c0a:  OK
	511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158:  OK
	dockertarsum: WARNING: 1 computed checksums did NOT match
	$ echo $?
	1

Additionally, if there are hashes for layers in the being checked, that are not
present in the image layers being read in, they will show with result of "NOT
FOUND", but will not cause a non-zero exit code. This is for the use-case that
there were sums calculated on a number of images, but a lesser set is being
validated presently.
Example:

	$ dockertarsum -c ./busybox.tar.sums ./busybox.tar 
	120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16:  OK
	42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229:  OK
	511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158:  OK
	a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721:  OK
	deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef:  NOT FOUND
	$ echo $?
	0

### Root Filesystem tars

There may be times where it is needed to calculate the fixed time of a tar
archive, that is not a tar of layers and would not include the json metadata
payload.

For this there is the `-r` flag.

	$ docker save busybox | tar xO 120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16/layer.tar | dockertarsum -r
	tarsum+sha256:cea0d2071b01b0a79aa4a05ea56ab6fdf3fafa03369d9f4eea8d46ea33c43e5f  -:-


### Detached Signatures

Here is a short screencast on the workflow for a detached GPG signature
workflow and validating an images' checksums.

![screenshot](http://vbatts.fedorapeople.org/docker/dockertarsum.png)
Video: [detached signature image workflow](http://vbatts.fedorapeople.org/docker/docker_image_signature-vbatts-07012014.webm)

The over of the commands for this process are:

	$ go get github.com/vbatts/docker-utils/cmd/dockertarum                         
	$ docker save -o busybox.tar busybox                                            
	$ dockertarsum busybox.tar > busybox.image.sums                                 
	$ cat busybox.image.sums                                                        
	$ dockertarsum -c ./busybox.image.sums ./busybox.tar                            
	$ echo $?                                                                       
	$ gpg -ba ./busybox.image.sums                                                  
	$ cat busybox.images.sum.asc                                                    
	$ gpg --verify busybox.images.sum.asc                                           

## d2r

Tooling for a static v1 Docker registry. This is useful for serving a read-only
registry.

### Installing

	go get github.com/vbatts/docker-utils/cmd/d2r

### Usage

	$ docker save busybox | d2r -o ./static -
  $ find ./static -type f
  ./static/v1/repositories/busybox/images
  ./static/v1/repositories/busybox/tags
  ./static/v1/_ping
  ./static/v1/images/120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16/layer
  ./static/v1/images/120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16/tarsum
  ./static/v1/images/120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16/json
  ./static/v1/images/a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721/layer
  ./static/v1/images/a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721/tarsum
  ./static/v1/images/a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721/json
  ./static/v1/images/a9eb172552348a9a49180694790b33a1097f546456d041b6e82e4d7716ddb721/ancestry
  ./static/v1/images/42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229/layer
  ./static/v1/images/42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229/tarsum
  ./static/v1/images/42eed7f1bf2ac3f1610c5e616d2ab1ee9c7290234240388d6297bc0f32c34229/json
  ./static/v1/images/511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/layer
  ./static/v1/images/511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/tarsum
  ./static/v1/images/511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/json


# Contributing

Pull requests? Yes, please

# License

See LICENSE
