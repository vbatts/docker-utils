dockertarsum image
----

An image for use in checking image checksums


Usage
===

	docker save <IMAGE> | docker run -i vbatts/dockertarsum


Example
===

	$> docker save busybox | docker run -i vbatts/dockertarsum
	dockertarsum: OK
	tarsum+sha256:d1f91d2d9510d90126eb4de06f8b27d955613ae90ad6570c6e66449c4e11a662  -:511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158
	tarsum+sha256:1450194b2e44ae373baa37a8b111ed7432cfefe25961477e5c3fa3b62b8ad21d  -:df7546f9f060a2268024c8a230d8639878585defcc1bc6f79d2728a13957871b
	tarsum+sha256:92397712012b61984226a9bc53c603a45c5dc37f5a22a1dbac996b9f44c60ae2  -:e433a6c5b276a31aa38bf6eaba9cd1cfd69ea33f706ed72b3f20bafde5cd8644
	tarsum+sha256:bd5999d1fa91e7c7ce774b09cc94fa706c325f070756f865fcad939697a6700a  -:e72ac664f4f0c6a061ac4ef332557a70d69b0c624b6add35f1c181ff7fff2287


Building
===

By default this image builds named for your user (`$USER`), and to sign the
binary with my own gpg key (`733F1362`).

To use your own key, run as:

	make DEFAULT_KEY=12312312

To clean and run through the whole process, and ensure checksum, run:

	make clean check

