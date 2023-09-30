const Input = require("@aux4/input");
const JsonIndex = require("../../lib/JsonIndex");

async function indexExecutor(params) {
  const path = await params.path;
  const jsonInput = await Input.readAsJson(path);

  const id = await params.id;
  if (!id) {
    throw new Error("parameter id is required");
  }

  const index = JsonIndex.index(jsonInput, id);
  console.log(JSON.stringify(index, null, 2));
}

module.exports = { indexExecutor };
