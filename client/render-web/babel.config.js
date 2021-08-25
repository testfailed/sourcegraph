// @ts-check

const { babelPresetEnvCommonOptions } = require('../../babel.config')
const jestConfig = require('../../jest.config.base')

if (!jestConfig.transformIgnorePatterns) {
  throw new Error('expected jest config to specify transformIgnorePatterns')
}

/** @type {import('@babel/core').TransformOptions} */
const config = {
  extends: '../../babel.config.js',
  ignore: jestConfig.transformIgnorePatterns.map(pattern => new RegExp(pattern)),
  presets: [
    [
      '@babel/preset-env',
      {
        // This program is run with Node instead of in the browser, so we need to compile it to
        // CommonJS.
        modules: 'commonjs',
        targets: {
          node: 'current'
        },
        ignoreBrowserslistConfig: true,
        ...babelPresetEnvCommonOptions,
      },
    ],
  ],
  plugins: [
    ['css-modules-transform', {
      // TODO(sqs): sync up with webpack.config.js localIdentName TODO(sqs): removed `_[hash:base64:5]`
      generateScopedName: '[name]__[local]',
      extensions: ['.css', '.scss'],
      camelCase: true,
    }],
  ]
}

module.exports = config
