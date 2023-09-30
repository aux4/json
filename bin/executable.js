#!/usr/bin/env node

const { Engine } = require("@aux4/engine");
const { indexExecutor } = require("./command/IndexExecutor");
const { collectExecutor } = require("./command/CollectExecutor");
const FormatExecutor = require("./command/FormatExecutor");
const { mergeExecutor } = require("./command/MergeExecutor");
const { groupExecutor } = require("./command/GroupExecutor");

const config = {
  profiles: [
    {
      name: "main",
      commands: [
        {
          name: "index",
          execute: indexExecutor,
          help: {
            text: "index json file",
            variables: [
              {
                name: "id",
                text: "id fields separated by comma"
              },
              {
                name: "path",
                text: "json file path",
                default: "$"
              }
            ]
          }
        },
        {
          name: "group",
          execute: groupExecutor,
          help: {
            text: "group json file by id",
            variables: [
              {
                name: "id",
                text: "id fields separated by comma"
              },
              {
                name: "path",
                text: "json file path",
                default: "$"
              }
            ]
          }
        },
        {
          name: "merge",
          execute: mergeExecutor,
          help: {
            text: "<files> merge json files",
            variables: [
              {
                name: "id",
                text: "id fields separated by comma"
              }
            ]
          }
        },
        {
          name: "collect",
          execute: collectExecutor,
          help: {
            text: "collect json from stream to array",
            variables: [
              {
                name: "path",
                text: "json root path",
                default: "$"
              }
            ]
          }
        },
        {
          name: "get",
          execute: FormatExecutor.get,
          help: {
            text: "<path> get json by path"
          }
        },
        {
          name: "pretty",
          execute: FormatExecutor.pretty,
          help: {
            text: "prettify json",
            variables: [
              {
                name: "path",
                text: "json root path",
                default: "$"
              }
            ]
          }
        },
        {
          name: "inline",
          execute: FormatExecutor.inline,
          help: {
            text: "inline json",
            variables: [
              {
                name: "path",
                text: "json root path",
                default: "$"
              }
            ]
          }
        }
      ]
    }
  ]
};

(async () => {
  const engine = new Engine({ aux4: config });

  const args = process.argv.splice(2);

  try {
    await engine.run(args);
  } catch (e) {
    console.error(e.message.red);
    process.exit(1);
  }
})();
