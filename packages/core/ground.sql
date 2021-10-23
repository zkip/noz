-- with recursive get_descendants as (
-- 	select ancestor, descendant from tHierarchy
-- 	union all
-- 	select t.* from get_descendants inner join get_descendants gd on t.descendant = 0
-- )

-- select * from get_descendants
-- select a.* from tHierarchy a where a.descendant in (select descendant from tHierarchy where ancestor=2)

-- delete from tHierarchy where descendant in (select * from (select descendant from tHierarchy) as tmp)
-- delete from tHierarchy where descendant in (select * from (select descendant from tHierarchy where ancestor=2) as _)

-- insert into tHierarchy( ancestor, descendant, distance, userPRI ) ( select ancestor, 1001, distance + 1, "sdf"  from tHierarchy where descendant = 0 )
-- delete from tHierarchy where targetPRI = "us/2" and descendant in (select * from (select descendant from tHierarchy where ancestor = "58d2959997f3fbb4d93eca7a2afea3cc") as _)

-- select descendant from tHierarchy where ancestor = "58d2959997f3fbb4d93eca7a2afea3cc"
-- select name from tHierarchyData where hierarchyID in (select ancestor from tHierarchy where descendant = "e1c65e2f3e00f5f596fce061f3cf699a" order by distance desc)
-- select name from tHierarchyData where hierarchyID in (select descendant from tHierarchy where ancestor = "e1c65e2f3e00f5f596fce061f3cf699a")

-- update tHierarchyData set
-- `order` = (
-- 	case when 78 < (select * from (select size from tHierarchyData where hierarchyID = "fff") a)
-- 	then 3
-- 	else 4 end
-- )
-- where id = 1

-- update tHierarchyData set
-- `order` = (
-- 	GREATEST(0, LEAST((select * from (select size from tHierarchyData where hierarchyID = "fff") a), 1234))
-- )
-- where id = 1

-- update ddd set bbc = NULL;

-- update ddd set bbc = IFNULL(bbc, "24");

-- select name from tHierarchyData where hierarchyID in (select descendant from tHierarchy where ancestor = "2b364d778a6afd01739e87eb62ec99c1" and distance = 1)

-- select hierarchyID, size, `order`, name from tHierarchyData where targetPRI = "us/1"

-- select ancestor from tHierarchy where descendant = ? order by distance desc

select * from tHierarchy;
select * from tHierarchyData;
select ancestor from tHierarchy where descendant = "7e01aa0df9308b1c2df2eace25cb9778" order by distance desc;
-- get full tree
select d.hierarchyID, d.name, d.order, t.ancestor from tHierarchyData d inner join tHierarchy t on t.descendant = d.hierarchyID order by t.descendant, t.distance desc;

select d.hierarchyID, d.size, d.order, d.name, t.ancestor from tHierarchyData d inner join tHierarchy t on d.targetPRI = "us/1" and t.descendant = d.hierarchyID order by t.descendant, t.distance desc;

-- delete posterity
delete from tHierarchy where descendant in (select * from (select descendant from tHierarchy where ancestor = "57233609e9e8799ba7335febc1fafb44") a);
-- select descendant from tHierarchy where ancestor = "57233609e9e8799ba7335febc1fafb44";

delete from tHierarchy where  descendant in 
(select * from (select descendant from tHierarchy where ancestor = "725b986bbe41d1587f864e3531a14836") a);

select name, hierarchyID from tHierarchyData where hierarchyID in 
(select descendant from tHierarchy where ancestor = "3774c0df837e6b4ae4d81bbd12f657c0");
select name, hierarchyID from tHierarchyData where hierarchyID in 
(select descendant from tHierarchy where ancestor != "3774c0df837e6b4ae4d81bbd12f657c0");
select * from (select t.descendant from tHierarchy t where t.ancestor = "dff4747c058e985837eea99564ce6598") b;
select d.name, t.descendant from tHierarchy t join tHierarchyData d on t.ancestor = "dff4747c058e985837eea99564ce6598" and t.descendant = hierarchyID;

select ancestor, distance, descendant from tHierarchy where descendant = "fb3d8fda1b0277b2dafca80d6fa49beb";
select * from (
	select ancestor, distance, descendant from tHierarchy where ancestor = "fb3d8fda1b0277b2dafca80d6fa49beb"
) tmp 
join tHierarchyData d on tmp.descendant = hierarchyID;

-- "b7d7b4b00717dab40796a0a7d4e34b47"

select d.size, d.order, d.hierarchyID, d.name, t.distance, t.descendant from tHierarchyData d
join tHierarchy t on t.ancestor = d.hierarchyID and t.distance = 1
and (t.descendant = "c52d8ee3c10f8bdde9fefb303e4ad230" or t.descendant = "b7d7b4b00717dab40796a0a7d4e34b47");

select d.size, d.order, d.hierarchyID, t.ancestor from tHierarchyData d join tHierarchy t on t.descendant = d.hierarchyID and t.distance = 1 and d.hierarchyID = "b7d7b4b00717dab40796a0a7d4e34b47";

-- move
delete a from tHierarchy as a
join tHierarchy as d on a.descendant = d.descendant
left join tHierarchy as x
on x.ancestor = d.ancestor and x.descendant = a.ancestor
where d.ancestor = 'e6f3569befd69048c31cb62f8251eb2f' and x.ancestor is null;

insert into tHierarchy (ancestor, descendant, distance)
select supertree.ancestor, subtree.descendant,
supertree.distance+subtree.distance+1
from tHierarchy as supertree join tHierarchy as subtree
where subtree.ancestor = 'e6f3569befd69048c31cb62f8251eb2f'
and supertree.descendant = '4167590ae540840dd4b43908f9947cfa';


select d.name, t.descendant from tHierarchy t join tHierarchyData d on t.ancestor = "dff4747c058e985837eea99564ce6598" and t.descendant = hierarchyID;

select * from (
	select ancestor, distance from tHierarchy where descendant in (select descendant from tHierarchy where ancestor = "4d8b0223e9068d2b9c8e4f503e7ee58b")
) tmp 
join tHierarchyData d on tmp.ancestor = hierarchyID;
SELECT * FROM TreePaths AS a;

SELECT * FROM TreePaths AS a
JOIN TreePaths AS d ON a.descendant = d.descendant;

DELETE a FROM TreePaths AS a
JOIN TreePaths AS d ON a.descendant = d.descendant
LEFT JOIN TreePaths AS x
ON x.ancestor = d.ancestor AND x.descendant = a.ancestor
WHERE d.ancestor = 'D' AND x.ancestor IS NULL;

INSERT INTO TreePaths (ancestor, descendant, length)
SELECT supertree.ancestor, subtree.descendant,
supertree.length+subtree.length+1
FROM TreePaths AS supertree JOIN TreePaths AS subtree
WHERE subtree.ancestor = 'D'
AND supertree.descendant = 'B';

SELECT ancestor from TreePaths where descendant = "E" order by length desc;

SELECT a.* FROM TreePaths AS a;

SELECT a.* FROM TreePaths AS a
JOIN TreePaths AS d ON a.descendant = d.descendant
LEFT JOIN TreePaths AS x
ON x.ancestor = d.ancestor AND x.descendant = a.ancestor;

SELECT a.* FROM TreePaths AS a
JOIN TreePaths AS d ON a.descendant = d.descendant
LEFT JOIN TreePaths AS x
ON x.ancestor = d.ancestor AND x.descendant = a.ancestor
WHERE d.ancestor = 'D' AND x.ancestor IS NULL;