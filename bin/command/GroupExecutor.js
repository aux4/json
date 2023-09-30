const Input = require("@aux4/input");
const JsonGroup = require("../../lib/JsonGroup");

async function groupExecutor(params) {
  const path = await params.path;
  const jsonInput = await Input.readAsJson(path);

  const id = await params.id;
  if (!id) {
    throw new Error("parameter id is required");
  }

  const group = JsonGroup.group(jsonInput, id);
  console.log(JSON.stringify(group, null, 2));
}

module.exports = { groupExecutor };
