# Domínio de Autenticação

## Registro

O registro é o processo pelo qual um novo usuário cria uma conta no sistema. Ele envolve a coleta de informações básicas, como nome, email e senha, e a validação desses dados para garantir que sejam válidos e únicos.
O processo começa gerando o hash da senha do usuário usando argon2. Criamos o ID do usuário manualmente usando UUID v7 para garantir a unicidade. Em seguida, armazenamos as informações do usuário no banco de dados, incluindo o ID, email e hash da senha.
Em seguida criamos um código de recuperação (a tabela recoveries tem 2 tipos de código: recovery e verification) e enviamos um email de verificação para o usuário. O email contém um link que o usuário deve clicar para verificar sua conta.
Após o usuário clicar no link de verificação, o sistema valida o código de recuperação e, se for válido, marca a conta do usuário como verificada. Isso garante que apenas usuários com emails válidos possam acessar o sistema.

## Login

O login é o processo pelo qual um usuário existente acessa sua conta no sistema. Ele envolve a coleta do email e senha do usuário, a validação dessas credenciais e a criação de uma sessão autenticada.
O processo começa buscando o usuário no banco de dados usando o email fornecido. Se o usuário for encontrado, o sistema verifica se o usuário está verificado e verifica a senha fornecida comparando-a com o hash armazenado usando argon2. Se a senha for válida, o sistema cria uma sessão autenticada para o usuário, permitindo que ele acesse as funcionalidades protegidas do sistema. Essa sessão é enviada para o usuário nos cookies, permitindo que ele permaneça autenticado em futuras solicitações.
A sessão funciona como um refresh token, permitindo que o usuário permaneça autenticado por um período prolongado sem precisar fazer login novamente. Ao alcançar 30 dias sem atividade, a sessão expira e o usuário precisará fazer login novamente para obter uma nova sessão.
