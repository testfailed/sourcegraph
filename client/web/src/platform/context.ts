import { ObservableQuery } from '@apollo/client'
import { Observable } from 'rxjs'
import { map, publishReplay, refCount, shareReplay } from 'rxjs/operators'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { createExtensionHost } from '@sourcegraph/shared/src/api/extension/worker'
import { getDocumentNode, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { mutateSettings, updateSettings } from '@sourcegraph/shared/src/settings/edit'
import { gqlToCascade } from '@sourcegraph/shared/src/settings/settings'
import { createAggregateError, asError } from '@sourcegraph/shared/src/util/errors'
import { LocalStorageSubject } from '@sourcegraph/shared/src/util/LocalStorageSubject'
import {
    toPrettyBlobURL,
    RepoFile,
    UIPositionSpec,
    ViewStateSpec,
    RenderModeSpec,
    UIRangeSpec,
    appendSubtreeQueryParameter,
} from '@sourcegraph/shared/src/util/url'

import { getWebGraphQLClient, requestGraphQL } from '../backend/graphql'
import { ViewerSettingsResult, ViewerSettingsVariables } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

/**
 * Creates the {@link PlatformContext} for the web app.
 */
export function createPlatformContext(): PlatformContext {
    const settingsWatcherPromise = watchViewerSettingsQuery()

    // Wrap `settingsWatcher` into `Observable` to have one `unsubscribe` call for both observables.
    const updatedSettings = new Observable<GQL.ISettingsCascade>(subscriber => {
        // Get `settingsWatcher` and subscribe to `updatedSettings` observable to updates
        const queryObservablePromise = settingsWatcherPromise.then(settingsWatcher =>
            settingsWatcher.subscribe(
                ({ data, errors }) => {
                    if (!data?.viewerSettings) {
                        subscriber.error(createAggregateError(errors))
                    } else {
                        subscriber.next(data.viewerSettings as GQL.ISettingsCascade)
                    }
                },
                error => subscriber.error(error)
            )
        )

        return () => {
            subscriber.unsubscribe()
            queryObservablePromise
                .then(queryObserver => queryObserver.unsubscribe())
                .catch(error => console.error(error))
        }
    }).pipe(shareReplay(1))

    const context: PlatformContext = {
        settings: updatedSettings.pipe(map(gqlToCascade), publishReplay(1), refCount()),
        updateSettings: async (subject, edit) => {
            const settingsWatcher = await settingsWatcherPromise

            // Unauthenticated users can't update settings. (In the browser extension, they can update client
            // settings even when not authenticated. The difference in behavior in the web app vs. browser
            // extension is why this logic lives here and not in shared/.)
            if (!window.context.isAuthenticatedUser) {
                const editDescription =
                    typeof edit === 'string' ? 'edit settings' : 'update setting `' + edit.path.join('.') + '`'
                const url = new URL(window.context.externalURL)
                throw new Error(
                    `Unable to ${editDescription} because you are not signed in.` +
                        '\n\n' +
                        `[**Sign into Sourcegraph${
                            url.hostname === 'sourcegraph.com' ? '' : ` on ${url.host}`
                        }**](${`${url.href.replace(/\/$/, '')}/sign-in`})`
                )
            }

            try {
                await updateSettings(context, subject, edit, mutateSettings)
            } catch (error) {
                if (asError(error).message.includes('version mismatch')) {
                    // The user probably edited the settings in another tab, so
                    // try once more.
                    await settingsWatcher.refetch()
                    await updateSettings(context, subject, edit, mutateSettings)
                } else {
                    throw error
                }
            }

            // The error will be emitted to consumers from the `updatedSettings` observable.
            settingsWatcher.refetch().catch(error => console.error(error))
        },
        getGraphQLClient: getWebGraphQLClient,
        requestGraphQL: ({ request, variables }) => requestGraphQL(request, variables),
        forceUpdateTooltip: () => Tooltip.forceUpdate(),
        createExtensionHost: () => Promise.resolve(createExtensionHost()),
        urlToFile: toPrettyWebBlobURL,
        getScriptURLForExtension: () => undefined,
        sourcegraphURL: window.context.externalURL,
        clientApplication: 'sourcegraph',
        sideloadedExtensionURL: new LocalStorageSubject<string | null>('sideloadedExtensionURL', null),
        telemetryService: eventLogger,
    }

    return context
}

function toPrettyWebBlobURL(
    context: RepoFile &
        Partial<UIPositionSpec> &
        Partial<ViewStateSpec> &
        Partial<UIRangeSpec> &
        Partial<RenderModeSpec>
): string {
    return appendSubtreeQueryParameter(toPrettyBlobURL(context))
}

const settingsCascadeFragment = gql`
    fragment SettingsCascadeFields on SettingsCascade {
        subjects {
            __typename
            ... on Org {
                id
                name
                displayName
            }
            ... on User {
                id
                username
                displayName
            }
            ... on Site {
                id
                siteID
                allowSiteSettingsEdits
            }
            latestSettings {
                id
                contents
            }
            settingsURL
            viewerCanAdminister
        }
        final
    }
`

/**
 * Creates Apollo query watcher for the viewer's settings. Watcher is used instead of the one-time query because we
 * want to use cached response if it's available. Callers should use settingsRefreshes#next instead of calling
 * this function, to ensure that the result is propagated consistently throughout the app instead of only being
 * returned to the caller.
 */
async function watchViewerSettingsQuery(): Promise<ObservableQuery<ViewerSettingsResult, ViewerSettingsVariables>> {
    const graphQLClient = await getWebGraphQLClient()

    return graphQLClient.watchQuery<ViewerSettingsResult, ViewerSettingsVariables>({
        fetchPolicy: 'cache-and-network',
        query: getDocumentNode(gql`
            query ViewerSettings {
                viewerSettings {
                    ...SettingsCascadeFields
                }
            }
            ${settingsCascadeFragment}
        `),
    })
}
