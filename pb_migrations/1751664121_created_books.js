/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = new Collection({
    "createRule": "@request.auth.id != \"\"",
    "deleteRule": "@request.auth.collectionName = \"_superusers\"",
    "fields": [
      {
        "autogeneratePattern": "[a-z0-9]{15}",
        "hidden": false,
        "id": "text3208210256",
        "max": 15,
        "min": 15,
        "name": "id",
        "pattern": "^[a-z0-9]+$",
        "presentable": false,
        "primaryKey": true,
        "required": true,
        "system": true,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text724990059",
        "max": 0,
        "min": 0,
        "name": "title",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text3182418120",
        "max": 0,
        "min": 0,
        "name": "author",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": true,
        "system": false,
        "type": "text"
      },
      {
        "convertURLs": false,
        "hidden": false,
        "id": "editor1843675174",
        "maxSize": 0,
        "name": "description",
        "presentable": false,
        "required": false,
        "system": false,
        "type": "editor"
      },
      {
        "hidden": false,
        "id": "number3402113753",
        "max": null,
        "min": 0,
        "name": "price",
        "onlyInt": false,
        "presentable": false,
        "required": true,
        "system": false,
        "type": "number"
      },
      {
        "hidden": false,
        "id": "number1261852256",
        "max": null,
        "min": 0,
        "name": "stock",
        "onlyInt": false,
        "presentable": false,
        "required": true,
        "system": false,
        "type": "number"
      },
      {
        "hidden": false,
        "id": "select105650625",
        "maxSelect": 0,
        "name": "category",
        "presentable": false,
        "required": true,
        "system": false,
        "type": "select",
        "values": [
          "Fiction",
          "Non-Fiction",
          "Science",
          "History",
          "Biography",
          "Technology",
          "Art",
          "Children"
        ]
      },
      {
        "autogeneratePattern": "",
        "hidden": false,
        "id": "text3424449766",
        "max": 0,
        "min": 0,
        "name": "isbn",
        "pattern": "",
        "presentable": false,
        "primaryKey": false,
        "required": false,
        "system": false,
        "type": "text"
      },
      {
        "hidden": false,
        "id": "file484410058",
        "maxSelect": 1,
        "maxSize": 0,
        "mimeTypes": null,
        "name": "cover_image",
        "presentable": false,
        "protected": false,
        "required": false,
        "system": false,
        "thumbs": [
          "200x300"
        ],
        "type": "file"
      }
    ],
    "id": "pbc_2170393721",
    "indexes": [],
    "listRule": "",
    "name": "books",
    "system": false,
    "type": "base",
    "updateRule": "@request.auth.collectionName = \"_superusers\"",
    "viewRule": ""
  });

  return app.save(collection);
}, (app) => {
  const collection = app.findCollectionByNameOrId("pbc_2170393721");

  return app.delete(collection);
})
