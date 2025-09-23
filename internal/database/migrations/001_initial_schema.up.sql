CREATE TABLE IF NOT EXISTS employee (
  emp_id  varchar(20) not null,
  create_ts timestamptz NOT NULL DEFAULT NOW(),
  update_ts timestamptz NOT NULL DEFAULT NOW(),
  updated_by text not null DEFAULT '',
  name  varchar(500) not null,
  email_id text not null default '',
  mobile text not null default '',
  telegram_id  text not null default '',

  off_role boolean not null default 'false',
  dept_name  varchar(200) not null default '',
  designation  varchar(100) not null default '',
  grade  varchar(100) not null default '',

  manager_id varchar(20) REFERENCES employee(emp_id) ON DELETE SET NULL,
  deviation_approver varchar(20) REFERENCES employee(emp_id) ON DELETE SET NULL,

  active boolean not null default 'true',
  shift_group text not null,
  PRIMARY KEY (emp_id)
);

CREATE TABLE IF NOT EXISTS shift_def (
  shift_cd varchar(3) not null,
  shift_group varchar(10) not null,
  create_ts timestamptz NOT NULL DEFAULT NOW(),
  update_ts timestamptz NOT NULL DEFAULT NOW(),
  in_time_start timestamptz not null,
  in_time_end timestamptz not null,
  min_halfday_duration int not null,
  min_present_duration int not null,
  double_shift_allowed boolean not null default 'true',
  min_double_shift_duration int not null,
  PRIMARY KEY (shift_cd)
);

-- Inserts are idempotent
CREATE TABLE IF NOT EXISTS raw_punch (
  emp_id  varchar(20) not null,
  punch_ts timestamptz  not null,
  device varchar(10) not null default '',
  punch_type  varchar(4) not null default '', --IN/OUT/BOTH
  
  PRIMARY KEY (emp_id, punch_ts)
);

-- processed punch
CREATE TABLE IF NOT EXISTS punch (
  emp_id  varchar(20) not null REFERENCES employee(emp_id),
  punch_date  timestamptz  not null,

  sys_shift_cd varchar(3) REFERENCES shift_def(shift_cd) ON DELETE SET NULL,
  sys_double_shift boolean not null default 'false',
  
  final_shift_cd varchar(3)   REFERENCES shift_def(shift_cd) ON DELETE SET NULL,
  final_double_shift boolean not null default 'false',

  create_ts timestamp NOT NULL DEFAULT NOW(),
  update_ts timestamp NOT NULL DEFAULT NOW(),
  updated_by text not null,

  deviation BOOLEAN not null default 'false',
  deviation_status varchar(100) not null default '',
  deviation_decision_ts timestamptz not null DEFAULT '1900-01-01 00:00:00 UTC',
  deviation_decision_by varchar(20) not null default '',
  
-- prop wil include {[{in_ts, out_ts}]} after senetization of min duration, and max gap for lunch. 
  props jsonb NOT NULL DEFAULT '{}',
  
  PRIMARY KEY (emp_id, punch_date)
);

create table if not exists punch_deviation_axn (
  id   SERIAL PRIMARY KEY,
  emp_id  varchar(20) not null REFERENCES employee(emp_id),
  punch_date  timestamptz  not null,
  event_ts timestamptz not null default now(),
  axn_type varchar(10) not null,
  props jsonb NOT NULL DEFAULT '{}'
);

--roster can be updated for future dates only.
CREATE TABLE IF NOT EXISTS roster (
  emp_id  varchar(20) not null REFERENCES employee(emp_id),
  roster_date timestamptz  not null,
  shift_cd varchar(3) not null,

  create_ts timestamp NOT NULL DEFAULT NOW(),
  update_ts timestamp NOT NULL DEFAULT NOW(),

  props jsonb NOT NULL DEFAULT '{}',

  PRIMARY KEY (emp_id, roster_date)
);
