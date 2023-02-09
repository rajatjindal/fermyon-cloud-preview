const childProcess = require('child_process')
const os = require('os')
const process = require('process')

async function main() {
    const downloadPath = await tc.downloadTool(downloadURL);

    const mainScript = `${__dirname}/fermyon-cloud-preview`
    const spawnSyncReturns = childProcess.spawnSync(mainScript, { stdio: 'inherit' })
    const status = spawnSyncReturns.status
    if (typeof status === 'number') {
        process.exit(status)
    }
    process.exit(1)
}

main()