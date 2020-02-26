# API
## Objects
### Item
* **Description**: Содержит описание товара
* **Fields**:
  * **title (string)**: Наименование товара, может совпадать у нескольких товаров
  * **id (uint64)**: Уникальный индетификатор товара
  * **category (string)**: Категория товара, может совпадать у нескольких товаров

## Methods
### AddItem()
* **Description**: Добавляет предмет в список
* **HttpMethod**: POST
* **UrlPath**: /items
* **Input**: Item
* **Input-type**: application/json

### GetItem()
* **Description**: Возвращает указанный предмет по id
* **HttpMethod**: GET
* **UrlPath**: /items/{id}
* **Output**: Item
* **Output-type**: application/json

### DeleteItem()
* **Description**: Удаляет указанный предмет из списка
* **HttpMethod**: DELETE
* **UrlPath**: /items/{id}

### UpdateItem()
* **Description**: Обновляет указанный предмет
* **HttpMethod**: PUT
* **UrlPath**: /items/{id}
* **Input**: Item
* **Input-type**: application/json

### GetItemList()
* **Description**: Возвращает список всех предметов
* **HttpMethod**: GET
* **UrlPath**: /items
* **Url-Parameters**:
  * **title (string, optional)**: Фильтрует список, оставляя только предметы с указанным title. Если указано несколько, будут возвращены все предметы с указанными title.
  * **id (uint64, optional)**: Фильтрует список, оставляя только предметы с указанным id. Если указано несколько, будут возвращены все предметы с указанными id.
  * **category (string, optional)**: Фильтрует список, оставляя только предметы с указанным category. Если указано несколько, будут возвращены все предметы с указанными category.
* **Output**: array(Item)
* **Output-type**: application/json