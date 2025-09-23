
-- name: GetEmpByIds :many
SELECT * FROM employee
WHERE emp_id = ANY(@empIds::text[]);

-- name: GetEmpByEmails :many
SELECT * FROM employee
WHERE email_id = ANY(@emails::text[]);

-- name: GetEmpByTelegramIds :many
SELECT * FROM employee
WHERE telegram_id = ANY(@telegramIds::text[]);;

-- name: GetEmpByApproverId :many
SELECT * FROM employee
WHERE deviation_approver = $1;

-- name: GetEmpByNameLike :many
SELECT * FROM employee 
WHERE name ILIKE $1;

-- name: GetAllEmpForManagerId :many
WITH RECURSIVE subordinates AS (
    SELECT * FROM employee e
    WHERE e.emp_id = @managerID 

    UNION ALL
    -- Recursive member: Find subordinates of the current set of subordinates
    SELECT * FROM employee e
    INNER JOIN
        subordinates s ON e.manager_id = s.emp_id
)
SELECT *
FROM subordinates;


-- name: CreateEmp :one
INSERT INTO employee (
  emp_id,
  name,
  email_id,
  mobile,
  telegram_id,
  off_role,
  dept_name,
  designation,
  grade,
  manager_id,
  deviation_approver,
  shift_group
) VALUES (
  $1,
  $2, 
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
)
RETURNING *;

-- name: UpdateEmpStatus :one
WITH rows AS (
  UPDATE employee
    set active = $2
  WHERE emp_id = $1
    RETURNING 1
)
SELECT count(*) FROM rows;


-- name: ScoreAndTests :many
--SELECT sqlc.embed(students), sqlc.embed(test_scores)
--FROM students
--JOIN test_scores ON test_scores.student_id = students.id
--WHERE students.id = $1;