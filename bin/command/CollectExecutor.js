const Input = require("@aux4/input");
const JsonCollect = require("../../lib/JsonCollect");

async function collectExecutor(params) {
  const path = await params.path;
  const stream = Input.stream(path);

  const collection = await JsonCollect.collect(stream);
  console.log(JSON.stringify(collection, null, 2));
}

module.exports = { collectExecutor };
