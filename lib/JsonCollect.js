class JsonCollect {
  static async collect(stream) {
    return new Promise((resolve, reject) => {
      const list = [];

      stream.on("data", data => {
        list.push(data);
      });

      stream.on("error", err => reject(err));

      stream.on("end", () => {
        resolve(list);
      });
    });
  }
}

module.exports = JsonCollect;
