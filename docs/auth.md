# API
## Objects
### ErrorResponse
* **Description**: Объект, содержащий ошибку. Возвращается любым методом в случае ошибки.
* **Fields**:
  * **error (string)**: Текстовое описание возникшей ошибки.

## Methods
### SignUp()
* **Description**: Регистрация нового пользователя.
* **HttpMethod**: POST
* **UrlPath**: /signup
* **Input-type**: application/json
* **Input**:
  * **username (string)**: Уникальное имя пользователя.
  * **password (string)**: Пароль пользователя.
  * **phone_number (string)**: Номер телефона для подтверждения регистрации.

### SignIn()
* **Description**: Авторизация уже зарегистрированного пользователя.
* **HttpMethod**: POST
* **UrlPath**: /signin
* **Input-type**: application/json
* **Input**:
  * **username (string)**: Имя пользователя.
  * **password (string)**: Пароль пользователя.
* **Output-type**: application/json
* **Output**:
  * **access_token (string)**: Краткосрочный токен для аунтефикации в сервисах.
  * **refresh_token (string)**: Долгосрочный токен для обновления access токена.

### Validate()
* **Description**: Проверят авторизацию пользователя по заголовку *auth*
* **HttpMethod**: GET
* **UrlPath**: /validate
* **RpcMethod**: AuthRpc.Validate
* **Authorization**: required
* **Output-type**: application/json
* **Output**:
  * **username (string)**: Username пользователя.
  * **permissions (array(string))**: Права пользователя в системе.

### Refresh()
* **Description**: Обновляет access токен.
* **HttpMethod**: PUT
* **UrlPath**: /refresh
* **Authorization**: required
* **Input-type**: application/json
* **Input**:
  * **refresh_token (string)**: Токен обновления.
* **Output-type**: application/json
* **Output**:
  * **access_token (string)**: Новый access токен.

### SetPermissions()
* **Description**: Устанавливает новые права для указанного пользователя.
* **HttpMethod**: PUT
* **UrlPath**: /set_permissions
* **Authorization**: required
* **Input-type**: application/json
* **Input**:
  * **username (string)**: Username пользователя.
  * **permissions (array(string))**: Новые права пользователя в системе.
