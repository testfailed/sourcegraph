import { InMemoryCache } from '@apollo/client'

import { TypedTypePolicies } from '../graphql-operations'

import { IExtensionRegistry } from './schema'

// Defines how the Apollo cache interacts with our GraphQL schema.
// See https://www.apollographql.com/docs/react/caching/cache-configuration/#typepolicy-fields
const typePolicies: TypedTypePolicies = {
    ExtensionRegistry: {
        merge(existing: IExtensionRegistry, incoming: IExtensionRegistry): IExtensionRegistry {
            return incoming
        },
    },
}

export const cache = new InMemoryCache({
    typePolicies,
})
