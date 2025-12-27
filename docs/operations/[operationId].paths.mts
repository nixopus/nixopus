import spec from '../src/openapi.json' with { type: 'json' }

const httpVerbs = ['get', 'post', 'put', 'delete', 'patch', 'head', 'options']

function encodeOperationId(operationId: string): string {
    return operationId.replace(/\//g, '_').replace(/:/g, '-')
}

export default {
    paths() {
        const paths = (spec as any).paths || {}

        return Object.keys(paths)
            .flatMap((path: string) => {
                return httpVerbs
                    .filter((verb: string) => paths[path][verb])
                    .map((verb: string) => {
                        const { operationId, summary } = paths[path][verb]
                        return {
                            params: {
                                operationId: encodeOperationId(operationId),
                                originalOperationId: operationId,
                                pageTitle: `${summary} - ${operationId}`,
                            },
                        }
                    })
            })
    },
}
