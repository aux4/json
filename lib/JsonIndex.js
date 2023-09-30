class JsonIndex {
  static index(jsonInput, id) {
    if (!id) {
      throw new Error("parameter id is required");
    }

    const index = {};

    const data = Array.isArray(jsonInput) ? jsonInput : [jsonInput];

    data.forEach(item => {
      const itemId = JsonIndex.generateId(item, id);

      let existingItem = index[itemId];
      if (existingItem) {
        if (!Array.isArray(existingItem)) {
          existingItem = [existingItem];
        }
        existingItem.push(item);
      } else {
        existingItem = item;
      }

      index[itemId] = existingItem;
    });

    return index;
  }

  static generateId(object, id) {
    const fields = id.split(",");
    return fields.map(field => object[field]).join("|");
  }
}

module.exports = JsonIndex;
