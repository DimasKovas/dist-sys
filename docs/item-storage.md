# API
## Objects
### Item
* **Description**: Содержит описание товара.
* **Fields**:
  * **title (string)**: Наименование товара, может совпадать у нескольких товаров.
  * **id (uint64, optional)**: Уникальный индетификатор товара. Всегда присутсвует в ответах сервера, но допускается отсутствие в некоторых пользовательских запросах (см. описание запросов).
  * **category (string)**: Категория товара, может совпадать у нескольких товаров.
  * **universal_code (string)**: Универсальный код товара. Должен быть уникальным для всех товаров. При отсутсвии кода он будет вычислен автоматически как хеш от других полей структуры (кроме id). 

### ErrorResponse
* **Description**: Объект, содержащий ошибку. Возвращается любым методом в случае ошибки.
* **Fields**:
  * **error (string)**: Текстовое описание возникшей ошибки.

## Methods
### AddItem()
* **Description**: Добавляет предмет в список
* **HttpMethod**: POST
* **UrlPath**: /item
* **Authorization**: required
* **Input-type**: application/json
* **Input**: Item (id is ignored)
* **Output-type**: application/json
* **Output**:
  * **id (uint64)**: Уникальный индетификатор созданного товара.

### GetItem()
* **Description**: Возвращает указанный предмет по id
* **HttpMethod**: GET
* **UrlPath**: /item/{id}
* **Authorization**: required
* **Output-type**: application/json
* **Output**: Item

### DeleteItem()
* **Description**: Удаляет указанный предмет из списка
* **HttpMethod**: DELETE
* **UrlPath**: /item/{id}
* **Authorization**: required

### UpdateItem()
* **Description**: Обновляет указанный предмет
* **HttpMethod**: PUT
* **UrlPath**: /item/{id}
* **Authorization**: required
* **Input-type**: application/json
* **Input**: Item (id is ignored)

### GetItemList()
* **Description**: Возвращает список всех предметов в порядке возврастания id
* **HttpMethod**: GET
* **UrlPath**: /items
* **Authorization**: required
* **Url-Parameters**:
  * **offset (uint, optional)**: Смещение возвращаемых предметов от начала. Первые offset предметов не будут возвращены.
  * **limit (uint, optional)**: Ограничение на количество возвращаемых предметов. Будут возвращены только первые limit предметов.
  * **category (string, optional)**: Фильтрует список, оставляя только предметы с указанным category. Если указано несколько, будут возвращены все предметы с указанными category.
* **Output-type**: application/json
* **Output**:
  * **count (uint)**: Сумарное количество предметов в базе данных.
  * **items (array(Item))**: Список запрошенных предметов.