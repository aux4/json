const JsonIndex = require("./JsonIndex");

class JsonGroup {
  static group(jsonInput, id) {
    const index = JsonIndex.index(jsonInput, id);

    const ids = id.split(",");
    const grouped = [];

    Object.values(index).forEach(value => {
      const items = Array.isArray(value) ? value : [value];

      const groupItem = {};
      ids.forEach(id => {
        groupItem[id] = items[0][id];
      });

      items.forEach(item => {
        Object.entries(item)
          .filter(([key]) => !ids.includes(key))
          .forEach(([key, value]) => {
            groupItem[key] = groupItem[key] || [];
            if (!groupItem[key].includes(value)) {
              groupItem[key].push(value);
            }
          });
      });

      grouped.push(groupItem);
    });

    return grouped;
  }
}

module.exports = JsonGroup;
