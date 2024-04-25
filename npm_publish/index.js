#!/usr/bin/env node

const path = require('node:path');
const childProcess = require('node:child_process');
const fsPromises = require('node:fs/promises');

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
const ARCH_MAPPING = {
  x64: 'amd64',
  arm64: 'arm64',
};
const PLATFORM_MACOS = 'darwin';
// The "fat" macOS binary includes code for both x64 and arm64 architectures
const GO_ARCH_FAT = 'fat';
// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING = {
  [PLATFORM_MACOS]: PLATFORM_MACOS,
  linux: 'linux',
  win32: 'windows',
};
/** @type {string?} */
let fullPath = null;

/**
 * @return {Promise<string>}
 */
export async function findGoIosBinary() {
  if (fullPath) {
    // return the previously cached value
    return fullPath;
  }

  const binaryName = `ios${process.platform === 'win32' ? '.exe' : ''}`;
  const binaryRoot = path.join(__dirname, 'dist');
  const goPlatform = PLATFORM_MAPPING[process.platform];
  const goArch = goPlatform === PLATFORM_MACOS ? GO_ARCH_FAT : ARCH_MAPPING[process.arch];
  if (goPlatform && goArch) {
    const sysFolderName = `go-ios-${goPlatform}-${goArch}_${goPlatform}_${goArch}`;
    fullPath = path.join(binaryRoot, sysFolderName, binaryName);
    try {
      await fsPromises.access(fullPath, fsPromises.constants.R_OK);
    } catch (ign) {
      fullPath = null;
    }
  }
  if (!fullPath) {
    throw new Error(
      `There is no precompiled go-ios binary for ${process.platform}@${process.arch} at '${binaryRoot}'`
    );
  }
  return fullPath;
}

/**
 * @returns {Promise<void>}
 */
async function main() {
  const binaryPath = await findGoIosBinary();
  const child = childProcess.spawn(binaryPath, process.argv.slice(2), {
    cwd: process.cwd(),
    env: process.env,
    stdio: [process.stdin, process.stdout, process.stderr],
  });
  await new Promise((resolve, reject) => {
    child.once('error', reject);
    child.once('exit', (code) => process.exit(code));
  });
}


if (require.main === module) {
  (async () => await main())();
}
