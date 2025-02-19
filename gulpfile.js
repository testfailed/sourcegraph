// @ts-check

const gulp = require('gulp')
const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
  watchSchema,
  cssModulesTypings,
  watchCSSModulesTypings,
} = require('./client/shared/gulpfile')
const { webpack: webWebpack, developmentServer } = require('./client/web/gulpfile')

/**
 * Generates files needed for builds.
 */
const generate = gulp.parallel(schema, graphQlSchema, graphQlOperations, cssModulesTypings)

/**
 * Starts all watchers on schema files.
 */
const watchGenerators = gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations, watchCSSModulesTypings)

/**
 * Generates files needed for builds whenever files change.
 */
const watchGenerate = gulp.series(generate, watchGenerators)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webWebpack)

/**
 * Watches everything and rebuilds on file changes.
 */
const dev = gulp.series(generate, gulp.parallel(watchGenerators, developmentServer))

module.exports = {
  generate,
  watchGenerate,
  build,
  dev,
  schema,
  graphQlSchema,
  watchGraphQlSchema,
  graphQlOperations,
  watchGraphQlOperations,
}
