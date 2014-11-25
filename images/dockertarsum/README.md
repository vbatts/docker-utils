= dockertarsum image

An image for use in checking image checksums


== Usage

	docker save <IMAGE> | docker run -i vbatts/dockertarsum


== Building

By default this image builds named for your user (`$USER`), and to sign the
binary with my own gpg key (`733F1362`).

To use your own key, run as:

	make DEFAULT_KEY=12312312

To clean and run through the whole process, and ensure checksum, run:

	make clean check

