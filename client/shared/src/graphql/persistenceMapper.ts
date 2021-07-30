import { pick } from 'lodash'

import { IQuery } from './schema'

/**
 * Hardcoded names of the queries which will be persisted to the local storage.
 * After the implementation of the `persistLink` which will support `@persist` directive
 * hardcoded query names will be deprecated.
 */
export const QUERIES_TO_PERSIST: (keyof IQuery)[] = ['viewerSettings', 'extensionRegistry']
export const ROOT_QUERY_KEY = 'ROOT_QUERY'

export interface CacheReference {
    __ref: string
}

export interface CacheObject {
    ROOT_QUERY: Record<string, unknown>
    [cacheKey: string]: unknown
}

// Ensures that we persist data required only for `QUERIES_TO_PERSIST`. Everything else is ignored.
export const persistenceMapper = (data: string): Promise<string> => {
    const initialData = JSON.parse(data) as CacheObject
    const dataToPersist: Record<string, unknown> = {}

    function findNestedCacheReferences(entry: unknown): void {
        if (!entry) {
            return
        }

        if (Array.isArray(entry)) {
            for (const item of entry) {
                findNestedCacheReferences(item)
            }
        } else if (isCacheReference(entry)) {
            const referenceKey = entry.__ref

            dataToPersist[referenceKey] = initialData[referenceKey]
            findNestedCacheReferences(initialData[referenceKey])
        } else if (entry && typeof entry === 'object') {
            for (const item of Object.values(entry)) {
                findNestedCacheReferences(item)
            }
        }
    }

    // Add responses of the specified queries to the result object.
    dataToPersist[ROOT_QUERY_KEY] = pick(initialData[ROOT_QUERY_KEY], [...QUERIES_TO_PERSIST, '__typename'])

    /**
     * Go through nested fields of the persisted responses and add references used there to the result object.
     *
     * Example ROOT_QUERY: { viewerSettings: { user: { __ref: 'User:01' } }, 'User:01': { ... } }
     * 'User:01' should be persisted, to have a complete cached response to the `viewerSettings` query.
     */
    for (const queryName of QUERIES_TO_PERSIST) {
        findNestedCacheReferences(initialData[ROOT_QUERY_KEY][queryName])
    }

    return Promise.resolve(JSON.stringify(dataToPersist))
}

function isCacheReference(entry: any): entry is CacheReference {
    return Boolean(entry.__ref)
}
