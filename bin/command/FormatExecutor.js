const Input = require("@aux4/input");

class FormatExecutor {
  static async get(params, action) {
    const path = action[0];
    const jsonInput = await Input.readAsJson(path);
    console.log(JSON.stringify(jsonInput, null, 2));
  }

  static async pretty(params) {
    const path = await params.path;

    const jsonInput = await Input.readAsJson(path);
    console.log(JSON.stringify(jsonInput, null, 2));
  }

  static async inline(params) {
    const path = await params.path;

    const jsonInput = await Input.readAsJson(path);
    console.log(JSON.stringify(jsonInput, null, 0));
  }
}

module.exports = FormatExecutor;
