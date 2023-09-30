const { JsonIndex } = require("../index");
const fs = require("fs").promises;

class JsonMerge {
  static async merge(files, id) {
    if (files.length < 2) {
      throw new Error("at least 2 files are required");
    }

    if (!id) {
      throw new Error("parameter id is required");
    }

    const filesContent = await Promise.all(files.map(file => fs.readFile(file, { encoding: "utf8" }).then(JSON.parse)));
    const indexes = await Promise.all(filesContent.map(fileContent => JsonIndex.index(fileContent, id)));

    const merged = [];

    Object.entries(indexes[0]).forEach(([indexId, value]) => {
      if (Array.isArray(value)) {
        value.forEach(item => {
          merge(indexId, item, indexes.slice(1)).forEach(mergedItem => merged.push(mergedItem));
        });
      } else {
        merge(indexId, value, indexes.slice(1)).forEach(mergedItem => merged.push(mergedItem));
      }
    });

    return merged;
  }
}

function merge(indexId, item, indexes) {
  let mergedItems = [{ ...item }];

  indexes.forEach(index => {
    const found = index[indexId];
    if (!found) return;

    if (Array.isArray(found)) {
      mergedItems = found.flatMap(item => mergedItems.map(mergedItem => ({ ...mergedItem, ...item })));
    } else {
      mergedItems = mergedItems.map(mergedItem => ({ ...mergedItem, ...found }));
    }
  });

  return mergedItems;
}

module.exports = JsonMerge;
