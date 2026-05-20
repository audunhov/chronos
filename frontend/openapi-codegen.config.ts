import {
  generateSchemaTypes,
  generateFetchers,
} from "@openapi-codegen/typescript";
import { defineConfig } from "@openapi-codegen/cli";

export default defineConfig({
  chronos: {
    from: {
      source: "file",
      relativePath: "openapi.yaml",
    },
    outputDir: "src/api",
    to: async (context) => {
      const { schemasFiles } = await generateSchemaTypes(context, {
        filenamePrefix: "chronos",
      });
      await generateFetchers(context, {
        filenamePrefix: "chronos",
        schemasFiles,
      });
    },
  },
});

