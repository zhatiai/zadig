package mongodb

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	mongotool "github.com/koderover/zadig/pkg/tool/mongo"
)

type NewNewJobInfoColl struct {
	*mongo.Collection

	coll string
}

func NewJobInfoCollFromOther() *NewNewJobInfoColl {
	name := "new_job_info"
	return &NewNewJobInfoColl{
		Collection: mongotool.Database(config.MongoDatabase()).Collection(name),
		coll:       name,
	}
}

func (c *NewNewJobInfoColl) GetCollectionName() string {
	return c.coll
}

func (c *NewNewJobInfoColl) EnsureIndex(ctx context.Context) error {
	mod := []mongo.IndexModel{
		{
			Keys: bson.D{
				bson.E{Key: "product_name", Value: 1},
			},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys: bson.D{
				bson.E{Key: "start_time", Value: 1},
			},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys: bson.D{
				bson.E{Key: "type", Value: 1},
			},
			Options: options.Index().SetUnique(false),
		},
	}

	_, err := c.Indexes().CreateMany(ctx, mod)
	return err
}

func (c *NewNewJobInfoColl) Create(ctx context.Context, args *models.JobInfo) error {
	if args == nil {
		return errors.New("configuration management is nil")
	}

	_, err := c.InsertOne(ctx, args)
	return err
}

func (c *NewNewJobInfoColl) GetProductionDeployJobs(startTime, endTime int64, projectName string) ([]*models.JobInfo, error) {
	query := bson.M{}
	query["start_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	query["production"] = true
	// TODO: currently the only job type with production env update function is zadig-deploy
	// if we added production update for helm-deploy type job, we need to update this query
	query["type"] = config.JobZadigDeploy
	if len(projectName) != 0 {
		query["product_name"] = projectName
	}

	resp := make([]*models.JobInfo, 0)

	cursor, err := c.Find(context.Background(), query, options.Find())
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &resp)

	return resp, err
}

func (c *NewNewJobInfoColl) GetTestJobs(startTime, endTime int64, projectName string) ([]*models.JobInfo, error) {
	query := bson.M{}
	query["start_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	query["type"] = config.JobZadigTesting

	if len(projectName) != 0 {
		query["product_name"] = projectName
	}

	resp := make([]*models.JobInfo, 0)

	cursor, err := c.Find(context.Background(), query, options.Find())
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &resp)

	return resp, err
}

func (c *NewNewJobInfoColl) GetBuildJobs(startTime, endTime int64, projectName string) ([]*models.JobInfo, error) {
	query := bson.M{}
	query["start_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	query["type"] = config.JobZadigBuild

	if len(projectName) != 0 {
		query["product_name"] = projectName
	}

	resp := make([]*models.JobInfo, 0)

	cursor, err := c.Find(context.Background(), query, options.Find())
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &resp)

	return resp, err
}

func (c *NewNewJobInfoColl) GetDeployJobs(startTime, endTime int64, projectName string) ([]*models.JobInfo, error) {
	query := bson.M{}
	query["start_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	query["type"] = bson.M{"$in": []string{string(config.JobZadigDeploy), string(config.JobZadigHelmDeploy), string(config.JobDeploy)}}

	if len(projectName) != 0 {
		query["product_name"] = projectName
	}

	resp := make([]*models.JobInfo, 0)

	cursor, err := c.Find(context.Background(), query, options.Find())
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &resp)

	return resp, err
}

func (c *NewNewJobInfoColl) GetCoarseGrainedData(startTime, endTime int64, projectName []string) (*JobInfoCoarseGrainedData, error) {
	match := bson.M{"$match": bson.M{"start_time": bson.M{"$gte": startTime, "$lt": endTime}}}

	if projectName != nil && len(projectName) != 0 {
		match["$match"].(bson.M)["product_name"] = bson.M{"$in": projectName}
	}

	project := bson.M{
		"$addFields": bson.M{
			"month": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m",
					"date": bson.M{
						"$add": []interface{}{
							bson.M{
								"$multiply": []interface{}{"$start_time", 1000},
							},
							time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					},
				},
			},
		},
	}

	group := bson.M{
		"$group": bson.M{
			"_id": "$month",
			"build_count": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$type", config.JobZadigBuild}},
						1, // 为真时的结果表达式
						0, // 为假时的结果表达式
					},
				},
			},
			"test_count": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$type", config.JobZadigTesting}},
						1, // 为真时的结果表达式
						0, // 为假时的结果表达式
					},
				},
			},
			"deploy_count": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$in": []interface{}{"$type", config.JobZadigDeploy}},
						1, // 为真时的结果表达式
						0, // 为假时的结果表达式
					},
				},
			},
		},
	}

	after := bson.M{
		"$project": bson.M{
			"_id":          0,
			"month":        "$_id",
			"build_count":  1,
			"test_count":   1,
			"deploy_count": 1,
		},
	}

	pipeline := []bson.M{match, project, group, after}

	cursor, err := c.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	jobInfo := &JobInfoCoarseGrainedData{
		StartTime: startTime,
		EndTime:   endTime,
	}
	infos := make([]*MonthlyJobInfo, 0)
	err = cursor.All(context.Background(), &infos)
	if err != nil {
		return nil, err
	}
	jobInfo.MonthlyStat = infos
	return jobInfo, nil
}

func (c *NewNewJobInfoColl) GetJobInfos(startTime, endTime int64, projectName []string) ([]*models.JobInfo, error) {
	query := bson.M{
		"start_time": bson.M{"$gte": startTime, "$lt": endTime},
	}

	if projectName != nil && len(projectName) != 0 {
		query["product_name"] = bson.M{"$in": projectName}
	}

	resp := make([]*models.JobInfo, 0)
	cursor, err := c.Find(context.Background(), query)
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &resp)
	return resp, err
}

func (c *NewNewJobInfoColl) GetJobBuildTrendInfos(startTime, endTime int64, projectName []string) ([]*JobBuildTrendInfos, error) {
	match := bson.M{"$match": bson.M{"start_time": bson.M{"$gte": startTime, "$lt": endTime}}}
	if projectName != nil && len(projectName) != 0 {
		match["$match"].(bson.M)["product_name"] = bson.M{"$in": projectName}
	}

	group := bson.M{
		"$group": bson.M{
			"_id":       "$product_name",
			"documents": bson.M{"$push": "$value"},
		},
	}
	pipeline := []bson.M{match, group}

	cursor, err := c.Find(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}

	resp := make([]*JobBuildTrendInfos, 0)
	err = cursor.All(context.Background(), &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *NewNewJobInfoColl) GetBuildTrend(startTime, endTime int64, projectName []string) ([]*models.JobInfo, error) {
	query := bson.M{
		"start_time": bson.M{"$gte": startTime, "$lt": endTime},
		"type":       config.JobZadigBuild,
	}
	if projectName != nil && len(projectName) != 0 {
		query["product_name"] = bson.M{"$in": projectName}
	}

	resp := make([]*models.JobInfo, 0)
	cursor, err := c.Find(context.Background(), query)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &resp)
	if err != nil {
		return nil, err
	}

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].StartTime < resp[j].StartTime
	})
	return resp, nil
}

func (c *NewNewJobInfoColl) GetAllProjectNameByTypeName(startTime, endTime int64, typeName string) ([]string, error) {
	query := bson.M{}
	if startTime != 0 && endTime != 0 {
		query["start_time"] = bson.M{"$gte": startTime, "$lt": endTime}
	}
	if typeName != "" {
		query["type"] = typeName
	}

	distinct, err := c.Distinct(context.Background(), "product_name", query)
	if err != nil {
		return nil, err
	}

	resp := make([]string, 0)
	for _, v := range distinct {
		if name, ok := v.(string); ok {
			if name == "" {
				continue
			}
			resp = append(resp, name)
		}
	}
	return resp, nil
}

func (c *NewNewJobInfoColl) GetTestTrend(startTime, endTime int64, projectName []string) ([]*models.JobInfo, error) {
	query := bson.M{
		"start_time":   bson.M{"$gte": startTime, "$lt": endTime},
		"product_name": bson.M{"$in": projectName},
		"type":         config.JobZadigTesting,
	}

	resp := make([]*models.JobInfo, 0)
	cursor, err := c.Find(context.Background(), query)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &resp)
	if err != nil {
		return nil, err
	}

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].StartTime < resp[j].StartTime
	})
	return resp, nil
}

func (c *NewNewJobInfoColl) GetDeployTrend(startTime, endTime int64, projectName []string) ([]*models.JobInfo, error) {
	query := bson.M{
		"start_time":   bson.M{"$gte": startTime, "$lt": endTime},
		"product_name": bson.M{"$in": projectName},
		"type":         config.JobZadigDeploy,
	}

	resp := make([]*models.JobInfo, 0)
	cursor, err := c.Find(context.Background(), query)
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.Background(), &resp)
	if err != nil {
		return nil, err
	}

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].StartTime < resp[j].StartTime
	})
	return resp, nil
}
