const JsonMerge = require("../../lib/JsonMerge");

async function mergeExecutor(params, files) {
  const id = await params.id;
  if (!id) {
    throw new Error("parameter id is required");
  }

  const merge = await JsonMerge.merge(files, id);
  console.log(JSON.stringify(merge, null, 2));
}

module.exports = { mergeExecutor };
