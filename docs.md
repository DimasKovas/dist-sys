# API
## Objects
### Item
* **Description**: Содержит описание товара.
* **Fields**:
  * **title (string)**: Наименование товара, может совпадать у нескольких товаров.
  * **id (uint64, optional)**: Уникальный индетификатор товара. Всегда присутсвует в ответах сервера, но допускается отсутствие в некоторых пользовательских запросах (см. описание запросов).
  * **category (string)**: Категория товара, может совпадать у нескольких товаров.

### AddItemResponse
* **Description**: Объект, возвращаемый методом AddItem. Содержит описание созданного товара.
* **Fields**:
  * **id (uint64)**: Уникальный индетификатор созданного товара.

### ErrorResponse
* **Description**: Объект, содержащий ошибку. Возвращается любым методом в случае ошибки.
* **Fields**:
  * **error (string)**: Текстовое описание возникшей ошибки.

## Methods
### AddItem()
* **Description**: Добавляет предмет в список
* **HttpMethod**: POST
* **UrlPath**: /item
* **Input**: Item (id is ignored)
* **Input-type**: application/json
* **Output**: AddItemResponse
* **Output-type**: application/json

### GetItem()
* **Description**: Возвращает указанный предмет по id
* **HttpMethod**: GET
* **UrlPath**: /item/{id}
* **Output**: Item
* **Output-type**: application/json

### DeleteItem()
* **Description**: Удаляет указанный предмет из списка
* **HttpMethod**: DELETE
* **UrlPath**: /item/{id}

### UpdateItem()
* **Description**: Обновляет указанный предмет
* **HttpMethod**: PUT
* **UrlPath**: /item/{id}
* **Input**: Item (id is ignored)
* **Input-type**: application/json

### GetItemList()
* **Description**: Возвращает список всех предметов в порядке возврастания id
* **HttpMethod**: GET
* **UrlPath**: /items
* **Url-Parameters**:
  * **offset (uint, optional)**: Смещение возвращаемых предметов от начала. Первые offset предметов не будут возвращены.
  * **limit (uint, optional)**: Ограничение на количество возвращаемых предметов. Будут возвращены только первые limit предметов.
  * **category (string, optional)**: Фильтрует список, оставляя только предметы с указанным category. Если указано несколько, будут возвращены все предметы с указанными category.
* **Output**: array(Item)
* **Output-type**: application/json