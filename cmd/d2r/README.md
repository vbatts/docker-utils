d2r
===

Create a static docker-registry filesystem, from the output of a `docker save`

Supports updating existing registry and repostory:tags

Usage
=====

	$> d2r -h
	Usage of ./d2r: ./d2r [OPTIONS] <file.tar|->
	  (where '-' is from stdin)
	  -o="./static/": directory to land the output registry files
	  -v=false: show version

Building
========

	$> go get github.com/vbatts/d2r

or

	$> git clone git://github.com/vbatts/d2r
	$> cd d2r
	$> make

Example
=======

	$ docker save fedora | d2r -o ./static/ - 
	Extracted Layer: 4a646dfe724f389db06873a6376faa94ffe46f96388ef9c467fa91309c59f21e [tarsum+sha256:88ad94bb9b70683d2c73b1d93f18518e7f38ed8d4346505181e71696a6e42429]
	Extracted Layer: 511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158 [tarsum+sha256:fc2555cf0a0c89b767bd6849fe4b664d3c48a0585d62362c627f43ac02fde8b6]
	Extracted Layer: 58394af373423902a1b97f209a31e3777932d9321ef10e64feaaa7b4df609cf9 [tarsum+sha256:adeaa1f621e896c9beba7802073b2f32dbbc73732ae7a13d912cc99805c76eda]
	Extracted Layer: 783f460032a82c15b584f92588fe3a4b748fb2d9012a0db46a65ad29878f9d3b [tarsum+sha256:cc7789bcfd5324fe34e9abf4ea9b75eee81432226942baad63b81f0b1eb8b5c7]
	Extracted Layer: 8abc22fbb04266308ff408ca61cb8f6f4244a59308f7efc64e54b08b496c58db [tarsum+sha256:2b8a383a1c7d62d89678a94dfbabbd4c82e201ddf7d3d4f1d160988ab16c44c6]
	Extracted Layer: 924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f [tarsum+sha256:c420ded8424aa96cb9531929ca6ac9cfc0f56b401f27c552c2a1b1631426495d]
	fedora
	  20 :: 924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f
	  heisenbug :: 924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f
	  latest :: 58394af373423902a1b97f209a31e3777932d9321ef10e64feaaa7b4df609cf9
	  rawhide :: 783f460032a82c15b584f92588fe3a4b748fb2d9012a0db46a65ad29878f9d3b
	$ find ./static/
	./static/
	./static/v1
	./static/v1/repositories
	./static/v1/repositories/fedora
	./static/v1/repositories/fedora/images
	./static/v1/repositories/fedora/tags
	./static/v1/_ping
	./static/v1/images
	./static/v1/images/8abc22fbb04266308ff408ca61cb8f6f4244a59308f7efc64e54b08b496c58db
	./static/v1/images/8abc22fbb04266308ff408ca61cb8f6f4244a59308f7efc64e54b08b496c58db/layer
	./static/v1/images/8abc22fbb04266308ff408ca61cb8f6f4244a59308f7efc64e54b08b496c58db/tarsum
	./static/v1/images/8abc22fbb04266308ff408ca61cb8f6f4244a59308f7efc64e54b08b496c58db/json
	./static/v1/images/4a646dfe724f389db06873a6376faa94ffe46f96388ef9c467fa91309c59f21e
	./static/v1/images/4a646dfe724f389db06873a6376faa94ffe46f96388ef9c467fa91309c59f21e/layer
	./static/v1/images/4a646dfe724f389db06873a6376faa94ffe46f96388ef9c467fa91309c59f21e/tarsum
	./static/v1/images/4a646dfe724f389db06873a6376faa94ffe46f96388ef9c467fa91309c59f21e/json
	./static/v1/images/924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f
	./static/v1/images/924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f/layer
	./static/v1/images/924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f/tarsum
	./static/v1/images/924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f/json
	./static/v1/images/924a401326b8f7fe65281647378b69f918aeb161678ec8a4209e074f4c5beb5f/ancestry
	[...]
	$ docker save debian | d2r -o ./static/ - 
	[...]


Testing
=======

The ./static/ (or whatever chosen) directory is servable as static content on
any webserver, though for convenience you can use the an included simple file
server from ./fsrv

	$> go get github.com/vbatts/d2r/fsrv
	$> fsrv ./static/

or

	$> git clone git://github.com/vbatts/d2r
	$> cd d2r
	$> make fsrv


TODO
====

* decouple when the saved output has a registry in the image name
* perhaps make the tarsum optional? ...

LICENSE
=======
Copyright (c) Vincent Batts <vbatts@hashbangbash.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
