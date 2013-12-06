
CREATE OR REPLACE FUNCTION set_created()
RETURNS TRIGGER AS $$
BEGIN
  NEW.created = now(); 
  NEW.modified = NEW.created;
  RETURN NEW;
END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION set_modified()
RETURNS TRIGGER AS $$
BEGIN
  NEW.modified = now(); 
  RETURN NEW;
END;
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION random_string(length integer) RETURNS text AS 
$$
declare
  chars text[] := '{0,1,2,3,4,5,6,7,8,9,A,B,C,D,E,F,G,H,I,J,K,L,M,N,O,P,Q,R,S,T,U,V,W,X,Y,Z,a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z}';
  result text := '';
  i integer := 0;
begin
  if length < 0 then
    raise exception 'Given length cannot be less than 0';
  end if;
  for i in 1..length loop
    result := result || chars[ceil(61 * random())];
  end loop;
  return result;
end;
$$ language plpgsql;

CREATE OR REPLACE FUNCTION set_keys()
RETURNS TRIGGER AS $$
BEGIN
  NEW.publickey = random_string(64);
  NEW.secretkey = random_string(64);
  RETURN NEW;
END;
$$ LANGUAGE PLPGSQL;

DROP TABLE IF EXISTS accounts;
CREATE TABLE accounts (
    pkey bigserial NOT NULL PRIMARY KEY,
    
    name text NOT NULL DEFAULT '',
    email text NOT NULL UNIQUE,
    passhash bytea NOT NULL,
    publickey text NOT NULL UNIQUE,
    secretkey text NOT NULL,
    
    --  common
    version smallint NOT NULL DEFAULT 0,
    -- values HSTORE,
    created TIMESTAMP NOT NULL,
    modified TIMESTAMP NOT NULL
);
CREATE TRIGGER set_tsadd BEFORE INSERT ON accounts FOR EACH ROW EXECUTE PROCEDURE set_created();
CREATE TRIGGER set_tsmod BEFORE UPDATE ON accounts FOR EACH ROW EXECUTE PROCEDURE set_modified();
CREATE TRIGGER set_apikey BEFORE INSERT ON accounts FOR EACH ROW EXECUTE PROCEDURE set_keys();
