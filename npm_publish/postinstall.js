#!/usr/bin/env node

"use strict";
// Thanks to author of https://github.com/sanathkr/go-npm, we were able to modify his code to work with private packages
var _typeof = typeof Symbol === "function" && typeof Symbol.iterator === "symbol" ? function (obj) { return typeof obj; } : function (obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; };

var path = require('path'),
    mkdirp = require('mkdirp'),
    fs = require('fs');

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
var ARCH_MAPPING = {
    "ia32": "386",
    "x64": "amd64",
    "arm": "arm",
    "arm64": "arm64"
};

// Mapping between Node's `process.platform` to Golang's
var PLATFORM_MAPPING = {
    "darwin": "darwin",
    "linux": "linux",
    "win32": "windows",
    "freebsd": "freebsd"
};

async function getInstallationPath() {

    // `npm bin` will output the path where binary files should be installed

    const value = null //await execShellCommand("npm bin -g");


    var dir = null;
    if (!value || value.length === 0) {

        // We couldn't infer path from `npm bin`. Let's try to get it from
        // Environment variables set by NPM when it runs.
        // npm_config_prefix points to NPM's installation directory where `bin` folder is available
        // Ex: /Users/foo/.nvm/versions/node/v4.3.0
        var env = process.env;
        if (env && env.npm_config_prefix) {
            dir = path.join(env.npm_config_prefix, "bin");
        }
    } else {
        dir = value.trim();
    }
    //throw (dir)
    ///Users/danielpaulus/.nvm/versions/node/v19.7.0/lib/node_modules/go-ios/node_modules/.bin
    await mkdirp(dir);
    return dir;
}

async function verifyAndPlaceBinary(binName, binPath, callback) {
    if (!fs.existsSync(path.join(binPath, binName))) return callback('Downloaded binary does not contain the binary specified in configuration - ' + binName);

    // Get installation path for executables under node
    const installationPath = await getInstallationPath();
    // Copy the executable to the path
    fs.rename(path.join(binPath, binName), path.join(installationPath, binName), (err) => {
        if (!err) {
            console.info("Installed cli successfully");
            callback(null);
        } else {
            callback(err);
        }
    });
}

function validateConfiguration(packageJson) {

    if (!packageJson.version) {
        return "'version' property must be specified";
    }

    if (!packageJson.goBinary || _typeof(packageJson.goBinary) !== "object") {
        return "'goBinary' property must be defined and be an object";
    }

    if (!packageJson.goBinary.name) {
        return "'name' property is necessary";
    }

    if (!packageJson.goBinary.path) {
        return "'path' property is necessary";
    }
}

function parsePackageJson() {
    if (process.arch !== "arm64" && process.platform !== "darwin") {
        if (!(process.arch in ARCH_MAPPING)) {
            console.error("Installation is not supported for this architecture: " + process.arch);
            return;
        }
    }

    if (!(process.platform in PLATFORM_MAPPING)) {
        console.error("Installation is not supported for this platform: " + process.platform);
        return;
    }

    var packageJsonPath = path.join(".", "package.json");
    if (!fs.existsSync(packageJsonPath)) {
        console.error("Unable to find package.json. " + "Please run this script at root of the package you want to be installed");
        return;
    }

    var packageJson = JSON.parse(fs.readFileSync(packageJsonPath));
    var error = validateConfiguration(packageJson);
    if (error && error.length > 0) {
        console.error("Invalid package.json: " + error);
        return;
    }

    // We have validated the config. It exists in all its glory
    var binName = packageJson.goBinary.name;
    var binPath = packageJson.goBinary.path;
    var version = packageJson.version;
    if (version[0] === 'v') version = version.substr(1); // strip the 'v' if necessary v0.0.1 => 0.0.1

    // Binary name on Windows has .exe suffix
    if (process.platform === "win32") {
        binName += ".exe";
    }


    return {
        binName: binName,
        binPath: binPath,
        version: version
    };
}

/**
 * Reads the configuration from application's package.json,
 * validates properties, copied the binary from the package and stores at
 * ./bin in the package's root. NPM already has support to install binary files
 * specific locations when invoked with "npm install -g"
 *
 *  See: https://docs.npmjs.com/files/package.json#bin
 */
var INVALID_INPUT = "Invalid inputs";
async function install(callback) {

    var opts = parsePackageJson();
    if (!opts) return callback(INVALID_INPUT);
    mkdirp.sync(opts.binPath);
    console.info(`Copying the relevant binary for your platform ${process.platform}`);
    let src = `./dist/go-ios-${PLATFORM_MAPPING[process.platform]}-${ARCH_MAPPING[process.arch]}_${PLATFORM_MAPPING[process.platform]}_${ARCH_MAPPING[process.arch]}/${opts.binName}`;
    if (process.arch === "arm64" && process.platform === "darwin") {
        console.log("using amd64 build on M1 mac")
        src = `./dist/go-ios-${process.platform}-amd64_${process.platform}_amd64/${opts.binName}`;
    }

    if (process.arch === "ia32" && process.platform === "w32") {
        src = `./dist/go-ios-${PLATFORM_MAPPING[process.platform]}-amd64_${PLATFORM_MAPPING[process.platform]}_amd64/${opts.binName}`;
    }

    if (PLATFORM_MAPPING[process.platform] === "windows") {
        let cmd = `copy ${src} ${opts.binPath}/${opts.binName}`
        cmd = cmd.replace(/\//g, "\\")
        await execShellCommand(cmd);
    } else {
        await execShellCommand(`cp ${src} ${opts.binPath}/${opts.binName}`);
    }

    await verifyAndPlaceBinary(opts.binName, opts.binPath, callback);
    console.log("\x1b[32m", "go-ios installed, run 'ios --help' for details\n\n")
}

async function uninstall(callback) {
    var opts = parsePackageJson();
    try {
        const installationPath = await getInstallationPath();
        fs.unlink(path.join(installationPath, opts.binName), (err) => {
            if (err) {
                return callback(err);
            }
        });
    } catch (ex) {
        // Ignore errors when deleting the file.
    }
    console.info("Uninstalled cli successfully");
    return callback(null);
}

// Parse command line arguments and call the right method
var actions = {
    "install": install,
    "uninstall": uninstall
};
/**
 * Executes a shell command and return it as a Promise.
 * @param cmd {string}
 * @return {Promise<string>}
 */
function execShellCommand(cmd) {
    const exec = require('child_process').exec;
    return new Promise((resolve, reject) => {
        exec(cmd, (error, stdout, stderr) => {
            if (error) {
                console.warn(error);
            }
            resolve(stdout ? stdout : stderr);
        });
    });
}

var argv = process.argv;
if (argv && argv.length > 2) {
    var cmd = process.argv[2];
    if (!actions[cmd]) {
        console.log("Invalid command. `install` and `uninstall` are the only supported commands");
        process.exit(1);
    }

    actions[cmd](function (err) {
        if (err) {
            console.error(err);
            process.exit(1);
        } else {
            process.exit(0);
        }
    });
}
