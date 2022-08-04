import lz4.block
import lz4.frame 
#f = open("/Users/danielpaulus/privaterepos/go-ios/ios/dtx_codec/fixtures/lz4block.bin", "rb")
f = open("/Users/danielpaulus/privaterepos/go-ios/chunk2", "rb")
compressed = f.read()
print(len(compressed))
comspize = 18117
uncompsize = 65536
last_uncompressed = b''
result = lz4.block.decompress(compressed, uncompsize, dict=last_uncompressed)


newFile = open("uncomp.bin", "wb")
# write to file
newFile.write(result)

#decompressed = lz4.frame.decompress(compressed)

#https://github.com/libyal/dtformats/blob/main/documentation/Apple%20Unified%20Logging%20and%20Activity%20Tracing%20formats.asciidoc#lz4_compressed_block

#https://github.com/ydkhatri/UnifiedLogReader/blob/4e7448a752863abb0a5284c8499c3c13db64d103/UnifiedLog/Lib.py#L117