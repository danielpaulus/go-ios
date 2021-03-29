import AppKit
func createArchiveFiles(objects: NSArray, filename: String) {
    do {
        let archiver = NSKeyedArchiver(requiringSecureCoding: false)
        archiver.outputFormat = .xml
        for object in objects {
            archiver.encode(object)
        }
        archiver.finishEncoding()
        // let archivedXML = String(bytes: archiver.encodedData, encoding: .ascii)!

        let bin_archiver = NSKeyedArchiver(requiringSecureCoding: false)
        for object in objects {
            bin_archiver.encode(object)
        }
        bin_archiver.finishEncoding()

        let fullPath = URL(fileURLWithPath: filename + ".bin")
        try bin_archiver.encodedData.write(to: fullPath)
        let fullPath1 = URL(fileURLWithPath: filename + ".xml")
        try archiver.encodedData.write(to: fullPath1)
    } catch {
        print("Couldn't write file")
    }
}

let primitives: NSArray = [UInt64(1), UInt32(1), Double(1.0), 1.5, Data(base64Encoded: "YXNkZmFzZGZhZHNmYWRzZg==")!, true,
                           "Hello, World!", "Hello, World!", "Hello, World!", false, false, 42]
let mutableArray: NSMutableArray = [true, "Hello, World!", 42]
let mutableSet: NSMutableSet = [true, "Hello, World!", 42]
let nsset: NSSet = [true]
let nestedNsset: NSArray = [nsset, mutableSet]

let dict: NSDictionary = ["int": 1, "string": "string", "array": mutableArray]

createArchiveFiles(objects: primitives, filename: "../fixtures/primitives")
createArchiveFiles(objects: [primitives, mutableArray, nsset, mutableSet], filename: "../fixtures/arrays")
createArchiveFiles(objects: [true], filename: "../fixtures/onevalue")
createArchiveFiles(objects: [nestedNsset], filename: "../fixtures/nestedarrays")
createArchiveFiles(objects: [dict], filename: "../fixtures/dict")
