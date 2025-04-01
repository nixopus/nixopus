CREATE OR REPLACE FUNCTION notify_application_change() RETURNS trigger AS $$
DECLARE
  notification json;
BEGIN
  IF TG_TABLE_NAME = 'applications' THEN
    IF (TG_OP = 'DELETE') THEN
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'application_id', OLD.id,
        'data', row_to_json(OLD)
      );
    ELSE
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'application_id', NEW.id,
        'data', row_to_json(NEW)
      );
    END IF;
  ELSIF TG_TABLE_NAME = 'application_deployment_status' THEN
    IF (TG_OP = 'DELETE') THEN
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'id', OLD.id,
        'application_deployment_id', OLD.application_deployment_id,
        'data', row_to_json(OLD)
      );
    ELSE
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'id', NEW.id,
        'application_deployment_id', NEW.application_deployment_id,
        'data', row_to_json(NEW)
      );
    END IF;
  ELSIF TG_TABLE_NAME = 'application_deployment' THEN
    IF (TG_OP = 'DELETE') THEN
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'id', OLD.id,
        'application_id', OLD.application_id,
        'data', row_to_json(OLD)
      );
    ELSE
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'id', NEW.id,
        'application_id', NEW.application_id,
        'data', row_to_json(NEW)
      );
    END IF;
  ELSE
    IF (TG_OP = 'DELETE') THEN
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'id', OLD.id,
        'application_id', OLD.application_id,
        'data', row_to_json(OLD)
      );
    ELSE
      notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'id', NEW.id,
        'application_id', NEW.application_id,
        'data', row_to_json(NEW)
      );
    END IF;
  END IF;
  
  PERFORM pg_notify('application_changes', notification::text);
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER applications_notify 
AFTER INSERT OR UPDATE OR DELETE ON applications 
FOR EACH ROW EXECUTE FUNCTION notify_application_change();

CREATE TRIGGER application_status_notify 
AFTER INSERT OR UPDATE OR DELETE ON application_status 
FOR EACH ROW EXECUTE FUNCTION notify_application_change();

CREATE TRIGGER application_logs_notify 
AFTER INSERT OR UPDATE OR DELETE ON application_logs 
FOR EACH ROW EXECUTE FUNCTION notify_application_change();

CREATE TRIGGER application_deployment_notify 
AFTER INSERT OR UPDATE OR DELETE ON application_deployment 
FOR EACH ROW EXECUTE FUNCTION notify_application_change();

CREATE TRIGGER application_deployment_status_notify 
AFTER INSERT OR UPDATE OR DELETE ON application_deployment_status 
FOR EACH ROW EXECUTE FUNCTION notify_application_change(); 