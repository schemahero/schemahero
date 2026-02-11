create table department (
  id char(36) primary key not null,
  name varchar(255) not null
);

create table employee (
  id char(36) primary key not null,
  email varchar(255) not null
);

create table assignment (
  id char(36) not null,
  employee_id char(36) not null,
  department_id char(36) not null,
  assigned_at timestamptz not null default now(),
  created_at timestamptz not null default now(),
  primary key (id),
  constraint assignment_employee_id_fkey foreign key (employee_id) references employee (id) on delete cascade,
  constraint assignment_department_id_fkey foreign key (department_id) references department (id) on delete cascade
);

create index idx_assignment_employee_id on assignment (employee_id);
create unique index idx_assignment_unique on assignment (employee_id, department_id);
