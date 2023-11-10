//go:build ignore

/*
xcrun clang -DDECODE -fobjc-arc -fmodules -F /System/Library/PrivateFrameworks/ -framework CoreUtils main.m --output decode
xcrun clang -DENCODE -fobjc-arc -fmodules -F /System/Library/PrivateFrameworks/ -framework CoreUtils main.m --output encode
*/

#import <Foundation/Foundation.h>
#import <Foundation/NSJSONSerialization.h>

CFMutableDataRef OPACKEncoderCreateData(NSObject *obj, int32_t flags, int32_t *error);
NSObject* OPACKDecodeBytes(const void *ptr, size_t length, int32_t flags, int32_t *error);

NSData *encodeOPack(NSObject *obj) {
    int32_t err = 0;
    CFDataRef data = OPACKEncoderCreateData(obj, 0, &err);
    if (err != 0) {
        return nil;
    }
    return (__bridge NSData*)data;
}

NSObject *decodeOPack(NSData *d) {
    const void *p = [d bytes];
    size_t l = [d length];
    return OPACKDecodeBytes(p, l, 0, nil);
}

NSString *encodeBase64(NSData *d) {
    return [d base64EncodedStringWithOptions:0];
}

NSData *decodeBase64(NSString *input) {
    return [[NSData alloc] initWithBase64EncodedString:[input stringByTrimmingCharactersInSet:[NSCharacterSet whitespaceCharacterSet]] options:NSDataBase64DecodingIgnoreUnknownCharacters];
}

NSObject *parseJson(NSString *input) {
    NSError *error = nil;
    id res = [NSJSONSerialization JSONObjectWithData:[input dataUsingEncoding:NSASCIIStringEncoding] options:0 error:&error];
    return res;
}

NSString *encodeJson(NSObject *obj) {
    NSData *d = [NSJSONSerialization dataWithJSONObject:obj options:0 error:nil];
    return [[NSString alloc] initWithData:d encoding:NSASCIIStringEncoding];
}

NSString *encodePlist(NSObject *obj) {
    NSData *d = [NSPropertyListSerialization dataWithPropertyList:obj format:NSPropertyListXMLFormat_v1_0 options:0 error:nil];
    return [[NSString alloc] initWithData:d encoding:NSASCIIStringEncoding];
}

NSObject *parsePlist(NSData *d) {
    return [NSPropertyListSerialization propertyListWithData:d options:0 format:nil error:nil];
}

int main(int argc, const char * argv[]) {
    @autoreleasepool {
        char input[4096];
//        scanf("%s", input);
        fgets(input, 4096, stdin);
        NSString *inputString = [[NSString alloc] initWithCString:input encoding:NSUTF8StringEncoding];
#ifdef ENCODE
        printf("%s", [encodeBase64(encodeOPack(parsePlist(decodeBase64(inputString)))) cStringUsingEncoding:NSASCIIStringEncoding]);
#endif
#ifdef DECODE
        printf("%se", [encodePlist(decodeOPack(decodeBase64(inputString))) cStringUsingEncoding:NSUTF8StringEncoding]);
#endif
    }
    return 0;
}
