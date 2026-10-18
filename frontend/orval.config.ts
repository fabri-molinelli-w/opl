import { defineConfig } from 'orval'

export default defineConfig({
  opl: {
    input: {
      target: '../api/openapi.yaml',
    },
    output: {
      mode: 'tags-split',
      target: 'src/api/gen/endpoints.ts',
      schemas: 'src/api/gen/model',
      client: 'react-query',
      httpClient: 'fetch',
      indexFiles: true,
      clean: true,
      override: {
        mutator: {
          path: 'src/api/http.ts',
          name: 'customInstance',
        },
      },
    },
  },
})
