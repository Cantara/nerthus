package database

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/cantara/nerthus/aws/security"
	"github.com/cantara/nerthus/aws/util"
	"github.com/cantara/nerthus/crypto"
)

type Database struct {
	Identifier string
	Database   string
	Password   string
	Name       string
	Scope      string
	ARN        string
	Endpoint   string
	group      security.Group
	rds        *rds.RDS
	created    bool
}

func NewDatabase(database, scope string, group security.Group, db *rds.RDS) (d Database, err error) {
	err = util.CheckRDSSession(db)
	if err != nil {
		return
	}
	d = Database{
		Database: database,
		Name:     fmt.Sprintf("%s-%s-db", scope, database),
		Scope:    scope,
		Password: crypto.GenRandBase32String(48),
		group:    group,
		rds:      db,
	}
	return
}

func (d *Database) Create() (arn string, err error) {
	// Specify the details of the instance that you want to create
	result, err := d.rds.CreateDBCluster(&rds.CreateDBClusterInput{
		BackupRetentionPeriod:   aws.Int64(7),
		AllocatedStorage:        aws.Int64(8),
		DBClusterIdentifier:     aws.String(d.Name),
		DBClusterInstanceClass:  aws.String("db.t3.micro"),
		DatabaseName:            aws.String(d.Database),
		Engine:                  aws.String("postgres"),
		EngineVersion:           aws.String("13.4"),
		MasterUserPassword:      aws.String(d.Password),
		MasterUsername:          aws.String(d.Database),
		Port:                    aws.Int64(5432),
		AutoMinorVersionUpgrade: aws.Bool(true),
		StorageEncrypted:        aws.Bool(true),
		PubliclyAccessible:      aws.Bool(false),
		Tags: []*rds.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(d.Name),
			},
			{
				Key:   aws.String("Scope"),
				Value: aws.String(d.Scope),
			},
		},
	})
	if err != nil {
		err = util.CreateError{
			Text: fmt.Sprintf("Could not create database with name %s.", d.Name),
			Err:  err,
		}
		return
	}
	d.ARN = aws.StringValue(result.DBCluster.DBClusterArn)
	d.Endpoint = aws.StringValue(result.DBCluster.Endpoint)
	arn = d.ARN
	d.created = true
	return
}

func (d *Database) Delete() (err error) {
	if !d.created {
		return
	}
	_, err = d.rds.DeleteDBCluster(&rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(d.Identifier),
	})
	if err != nil {
		return
	}
	return
}

/*
result, err := svc.CreateDBCluster(input)
if err != nil {
    if aerr, ok := err.(awserr.Error); ok {
        switch aerr.Code() {
        case rds.ErrCodeDBClusterAlreadyExistsFault:
            fmt.Println(rds.ErrCodeDBClusterAlreadyExistsFault, aerr.Error())
        case rds.ErrCodeInsufficientStorageClusterCapacityFault:
            fmt.Println(rds.ErrCodeInsufficientStorageClusterCapacityFault, aerr.Error())
        case rds.ErrCodeDBClusterQuotaExceededFault:
            fmt.Println(rds.ErrCodeDBClusterQuotaExceededFault, aerr.Error())
        case rds.ErrCodeStorageQuotaExceededFault:
            fmt.Println(rds.ErrCodeStorageQuotaExceededFault, aerr.Error())
        case rds.ErrCodeDBSubnetGroupNotFoundFault:
            fmt.Println(rds.ErrCodeDBSubnetGroupNotFoundFault, aerr.Error())
        case rds.ErrCodeInvalidVPCNetworkStateFault:
            fmt.Println(rds.ErrCodeInvalidVPCNetworkStateFault, aerr.Error())
        case rds.ErrCodeInvalidDBClusterStateFault:
            fmt.Println(rds.ErrCodeInvalidDBClusterStateFault, aerr.Error())
        case rds.ErrCodeInvalidDBSubnetGroupStateFault:
            fmt.Println(rds.ErrCodeInvalidDBSubnetGroupStateFault, aerr.Error())
        case rds.ErrCodeInvalidSubnet:
            fmt.Println(rds.ErrCodeInvalidSubnet, aerr.Error())
        case rds.ErrCodeInvalidDBInstanceStateFault:
            fmt.Println(rds.ErrCodeInvalidDBInstanceStateFault, aerr.Error())
        case rds.ErrCodeDBClusterParameterGroupNotFoundFault:
            fmt.Println(rds.ErrCodeDBClusterParameterGroupNotFoundFault, aerr.Error())
        case rds.ErrCodeKMSKeyNotAccessibleFault:
            fmt.Println(rds.ErrCodeKMSKeyNotAccessibleFault, aerr.Error())
        case rds.ErrCodeDBClusterNotFoundFault:
            fmt.Println(rds.ErrCodeDBClusterNotFoundFault, aerr.Error())
        case rds.ErrCodeDBInstanceNotFoundFault:
            fmt.Println(rds.ErrCodeDBInstanceNotFoundFault, aerr.Error())
        case rds.ErrCodeDBSubnetGroupDoesNotCoverEnoughAZs:
            fmt.Println(rds.ErrCodeDBSubnetGroupDoesNotCoverEnoughAZs, aerr.Error())
        case rds.ErrCodeGlobalClusterNotFoundFault:
            fmt.Println(rds.ErrCodeGlobalClusterNotFoundFault, aerr.Error())
        case rds.ErrCodeInvalidGlobalClusterStateFault:
            fmt.Println(rds.ErrCodeInvalidGlobalClusterStateFault, aerr.Error())
        case rds.ErrCodeDomainNotFoundFault:
            fmt.Println(rds.ErrCodeDomainNotFoundFault, aerr.Error())
        default:
            fmt.Println(aerr.Error())
        }
    } else {
        // Print the error, cast err to awserr.Error to get the Code and
        // Message from an error.
        fmt.Println(err.Error())
    }
    return
}
*/
