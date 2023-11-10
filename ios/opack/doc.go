/*
Package opack implements the serialization format OPACK from Apple (implemented in the CoreUtils.framework)

# It's an incomplete implementation that only encodes data

#import <Foundation/Foundation.h>
#import <Foundation/NSJSONSerialization.h>

CFMutableDataRef OPACKEncoderCreateData(NSObject *obj, int32_t flags, int32_t *error);
NSObject* OPACKDecodeBytes(const void *ptr, size_t length, int32_t flags, int32_t *error);

	int main(int argc, const char * argv[]) {
	    @autoreleasepool {
	        NSError *e = nil;
	        NSFileHandle *stdInFh = [NSFileHandle fileHandleWithStandardInput];
	        NSData *stdin = [stdInFh readDataToEndOfFile];

	        int decode_error = 0;
	        NSObject *decoded = OPACKDecodeBytes([stdin bytes], [stdin length], 0, &decode_error);
	        if (decode_error) {
	            NSLog(@"Failed to decode: %d", decode_error);
	            return -1;
	        }

	        NSLog(@"decoded: %@", decoded);
	        NSData *json = [NSJSONSerialization dataWithJSONObject: decoded options: NSJSONWritingPrettyPrinted error: &e];
	        if (e) {
	            NSLog(@"Failed to write JSON: %@", e);
	            return -1;
	        }

	        NSFileHandle *stdOut = [NSFileHandle fileHandleWithStandardOutput];
	        [stdOut writeData: json];
	    }
	    return 0;
	}
*/
package opack
