import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import signale from 'signale'

import { BUILD_OPTIONS } from '@sourcegraph/web/dev/esbuild/build'

const outputDirectory = path.join(__dirname, 'out')
const outputBundlePath = path.join(outputDirectory, 'prerender.js')

export const buildBundle = async (watch = false): Promise<void> => {
    await esbuild.build({
        ...BUILD_OPTIONS,
        entryPoints: ['src/serve.ts'],
        splitting: false,
        outdir: undefined,
        outfile: outputBundlePath,
        platform: 'node',
        external: ['node-fetch'],
        // mainFields: ['module', 'main'],
        format: 'iife',
        // watch: true,

        // TODO(sqs): Tree shaking is disabled in the default build options due to
        // https://github.com/evanw/esbuild/pull/1458, but that bug doesn't affect the prerender
        // bundle since it only affects the CSS. So, we can enable (i.e., un-disable) tree shaking
        // in this build.
        treeShaking: undefined,
    })
    const stat = await fs.promises.stat(outputBundlePath)
    signale.success(
        `Built bundle: ${path.relative(__dirname, outputBundlePath)} [${(stat.size / (1024 * 1024)).toFixed(1)}MB]`
    )
}

export const watchBundle = async (): Promise<void> => buildBundle(true)
