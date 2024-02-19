FROM ubuntu:latest
RUN apt-get update&& apt-get install -y git build-essential pkg-config autoconf automake libtool-bin python3 python-is-python3 libssl-dev libusb-1.0-0-dev
RUN git clone https://github.com/libimobiledevice/libplist
RUN git clone https://github.com/libimobiledevice/libusbmuxd
RUN git clone https://github.com/libimobiledevice/libimobiledevice
RUN git clone https://github.com/libimobiledevice/libimobiledevice-glue
RUN git clone https://github.com/libimobiledevice/usbmuxd.git
RUN cd libplist && ./autogen.sh && make && make install
RUN cd libimobiledevice-glue && ./autogen.sh && make && make install
RUN cd libusbmuxd && ./autogen.sh && make && make install
RUN cd libimobiledevice && ./autogen.sh && make && make install
RUN cd usbmuxd && ./autogen.sh && make && make install
#if you don't this, you will get this weird error:
# usbmuxd: error while loading shared libraries: libimobiledevice-glue-1.0.so.0: cannot open shared object file: No such file or directory
# I guess a reboot would also work, tricky with a container though
RUN ldconfig
RUN usbmuxd --version

