alter table issues add constraint renamed_fkey foreign key (project_id) references projects (id);
